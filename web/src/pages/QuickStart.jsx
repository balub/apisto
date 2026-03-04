import { useState, useEffect } from 'react'
import { CodeSnippet } from '../components/CodeSnippet'
import { api } from '../lib/api'

const ARDUINO_TEMPLATE = (token, host) => `#include <Apisto.h>
#include <DHT.h>

#define WIFI_SSID     "YourWiFiNetwork"
#define WIFI_PASSWORD "YourWiFiPassword"
#define DEVICE_TOKEN  "${token}"
#define SERVER_HOST   "${host}"

DHT dht(4, DHT22);
Apisto device(DEVICE_TOKEN, SERVER_HOST);

void setup() {
  Serial.begin(115200);
  dht.begin();
  device.begin(WIFI_SSID, WIFI_PASSWORD);
}

void loop() {
  device.loop();

  static unsigned long lastSend = 0;
  if (millis() - lastSend > 5000) {
    device.send("temperature", dht.readTemperature());
    device.send("humidity", dht.readHumidity());
    lastSend = millis();
  }
}`

const CURL_TEMPLATE = (token) => `curl -X POST http://localhost:8080/api/v1/devices/${token}/telemetry \\
  -H "Content-Type: application/json" \\
  -d '{"temperature": 24.5, "humidity": 60, "relay_on": true}'`

export function QuickStart() {
  const [projects, setProjects] = useState([])
  const [selectedProject, setSelectedProject] = useState('')
  const [devices, setDevices] = useState([])
  const [selectedDevice, setSelectedDevice] = useState(null)
  const serverHost = location.hostname || 'localhost'

  useEffect(() => {
    api.listProjects().then(p => { setProjects(p || []); if (p?.[0]) setSelectedProject(p[0].id) }).catch(() => {})
  }, [])

  useEffect(() => {
    if (!selectedProject) return
    api.listDevices(selectedProject).then(d => {
      setDevices(d || [])
      setSelectedDevice(d?.[0] || null)
    }).catch(() => {})
  }, [selectedProject])

  const token = selectedDevice?.token || 'YOUR_DEVICE_TOKEN'

  const steps = [
    {
      n: 1,
      title: 'Start the server',
      desc: 'Run the full stack with Docker Compose.',
      code: 'git clone https://github.com/balub/apisto.git\ncd apisto\ndocker compose up -d',
      lang: 'bash',
    },
    {
      n: 2,
      title: 'Create a project',
      desc: 'Use the UI above or curl.',
      code: `curl -X POST http://localhost:8080/api/v1/projects \\\n  -H "Content-Type: application/json" \\\n  -d '{"name": "My Home"}'`,
      lang: 'bash',
    },
    {
      n: 3,
      title: 'Add a device',
      desc: 'Each device gets a unique token. Keep it secret.',
      code: `curl -X POST http://localhost:8080/api/v1/projects/${selectedProject || 'PROJECT_ID'}/devices \\\n  -H "Content-Type: application/json" \\\n  -d '{"name": "Sensor 1"}'`,
      lang: 'bash',
    },
    {
      n: 4,
      title: 'Upload the sketch',
      desc: 'Install PubSubClient and ArduinoJson libraries, then flash your ESP32.',
      code: ARDUINO_TEMPLATE(token, serverHost),
      lang: 'cpp',
    },
    {
      n: 5,
      title: 'Or test with curl',
      desc: 'No hardware? Simulate a device with HTTP.',
      code: CURL_TEMPLATE(token),
      lang: 'bash',
    },
  ]

  return (
    <div className="min-h-screen p-8 max-w-3xl mx-auto">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-white">Quick Start</h1>
        <p className="text-zinc-500 mt-1">From zero to live dashboard in under 15 minutes</p>
      </div>

      {projects.length > 0 && (
        <div className="card mb-8">
          <div className="flex gap-4">
            <div className="flex-1">
              <label className="label">Project</label>
              <select className="input" value={selectedProject} onChange={e => setSelectedProject(e.target.value)}>
                {projects.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
              </select>
            </div>
            <div className="flex-1">
              <label className="label">Device</label>
              <select className="input" value={selectedDevice?.id || ''} onChange={e => setSelectedDevice(devices.find(d => d.id === e.target.value))}>
                <option value="">Select device...</option>
                {devices.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            </div>
          </div>
          {selectedDevice && (
            <div className="mt-3 text-xs text-zinc-500">
              Code snippets below are pre-filled with token for <span className="text-white">{selectedDevice.name}</span>
            </div>
          )}
        </div>
      )}

      <div className="space-y-6">
        {steps.map((step) => (
          <div key={step.n} className="flex gap-4">
            <div className="shrink-0 w-8 h-8 rounded-full bg-accent/10 border border-accent/20 flex items-center justify-center">
              <span className="text-accent text-sm font-bold">{step.n}</span>
            </div>
            <div className="flex-1">
              <h3 className="font-semibold text-white mb-1">{step.title}</h3>
              <p className="text-zinc-500 text-sm mb-3">{step.desc}</p>
              <CodeSnippet code={step.code} language={step.lang} />
            </div>
          </div>
        ))}
      </div>

      <div className="mt-12 card bg-accent/5 border-accent/20">
        <div className="text-accent font-semibold mb-2">Step 5: See your data →</div>
        <p className="text-zinc-400 text-sm">
          Open your{' '}
          <a href="/" className="text-accent hover:underline">project dashboard</a>{' '}
          — widgets appear automatically for each telemetry key your device sends.
        </p>
      </div>
    </div>
  )
}
