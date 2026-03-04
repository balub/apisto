package mqtt

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type TelemetryMessage struct {
	DeviceID string
	Token    string
	Payload  []byte
}

type Handlers struct {
	OnTelemetry func(ctx context.Context, deviceID string, payload []byte)
	OnStatus    func(ctx context.Context, deviceID string, payload []byte)
	OnCommandAck func(ctx context.Context, commandID string)
}

type Client struct {
	client   mqtt.Client
	handlers Handlers

	// Token → device ID cache with TTL
	tokenCache sync.Map // map[string]tokenCacheEntry

	// DB lookup function for token validation
	lookupToken func(ctx context.Context, token string) (string, error)

	// Buffered channel for async telemetry processing
	telemetryCh chan TelemetryMessage

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type tokenCacheEntry struct {
	deviceID  string
	expiresAt time.Time
}

const (
	tokenCacheTTL   = 5 * time.Minute
	telemetryChanBuf = 512
	workerCount      = 4
)

func NewClient(brokerURL, clientID string, lookupToken func(ctx context.Context, token string) (string, error), handlers Handlers) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Client{
		handlers:    handlers,
		lookupToken: lookupToken,
		telemetryCh: make(chan TelemetryMessage, telemetryChanBuf),
		ctx:         ctx,
		cancel:      cancel,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Printf("mqtt: connection lost: %v", err)
	})
	opts.SetOnConnectHandler(func(cl mqtt.Client) {
		log.Println("mqtt: connected, subscribing to topics")
		c.subscribe(cl)
	})

	c.client = mqtt.NewClient(opts)

	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		cancel()
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}

	// Start worker pool
	for i := 0; i < workerCount; i++ {
		c.wg.Add(1)
		go c.telemetryWorker()
	}

	return c, nil
}

func (c *Client) subscribe(cl mqtt.Client) {
	subs := map[string]byte{
		"apisto/+/telemetry":    1,
		"apisto/+/status":       1,
		"apisto/+/commands/ack": 1,
	}
	for topic, qos := range subs {
		if token := cl.Subscribe(topic, qos, c.messageHandler); token.Wait() && token.Error() != nil {
			log.Printf("mqtt: subscribe %s error: %v", topic, token.Error())
		}
	}
}

func (c *Client) messageHandler(_ mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	parts := strings.Split(topic, "/")
	// Expected: apisto/{token}/{type} or apisto/{token}/commands/ack
	if len(parts) < 3 {
		return
	}
	token := parts[1]
	msgType := parts[2]

	deviceID, err := c.resolveToken(c.ctx, token)
	if err != nil {
		log.Printf("mqtt: invalid token %s: %v", token, err)
		return
	}

	switch msgType {
	case "telemetry":
		select {
		case c.telemetryCh <- TelemetryMessage{DeviceID: deviceID, Token: token, Payload: msg.Payload()}:
		default:
			log.Printf("mqtt: telemetry channel full, dropping message for device %s", deviceID)
		}
	case "status":
		if c.handlers.OnStatus != nil {
			go c.handlers.OnStatus(c.ctx, deviceID, msg.Payload())
		}
	case "commands":
		if len(parts) == 4 && parts[3] == "ack" && c.handlers.OnCommandAck != nil {
			// Extract command ID from payload
			go c.handlers.OnCommandAck(c.ctx, string(msg.Payload()))
		}
	}
}

func (c *Client) telemetryWorker() {
	defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			// Drain remaining messages
			for {
				select {
				case msg := <-c.telemetryCh:
					if c.handlers.OnTelemetry != nil {
						c.handlers.OnTelemetry(context.Background(), msg.DeviceID, msg.Payload)
					}
				default:
					return
				}
			}
		case msg := <-c.telemetryCh:
			if c.handlers.OnTelemetry != nil {
				c.handlers.OnTelemetry(c.ctx, msg.DeviceID, msg.Payload)
			}
		}
	}
}

func (c *Client) resolveToken(ctx context.Context, token string) (string, error) {
	if entry, ok := c.tokenCache.Load(token); ok {
		e := entry.(tokenCacheEntry)
		if time.Now().Before(e.expiresAt) {
			return e.deviceID, nil
		}
		c.tokenCache.Delete(token)
	}

	deviceID, err := c.lookupToken(ctx, token)
	if err != nil {
		return "", err
	}

	c.tokenCache.Store(token, tokenCacheEntry{
		deviceID:  deviceID,
		expiresAt: time.Now().Add(tokenCacheTTL),
	})
	return deviceID, nil
}

// Publish sends a message to the given MQTT topic.
func (c *Client) Publish(topic string, payload []byte) error {
	token := c.client.Publish(topic, 1, false, payload)
	token.Wait()
	return token.Error()
}

// Disconnect gracefully shuts down the MQTT client.
func (c *Client) Disconnect() {
	c.cancel()
	c.wg.Wait()
	c.client.Disconnect(500)
}

// InvalidateToken removes a token from the cache (call when device is deleted).
func (c *Client) InvalidateToken(token string) {
	c.tokenCache.Delete(token)
}
