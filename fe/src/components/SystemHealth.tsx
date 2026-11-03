import React, { useState } from 'react';
import { Terminal } from 'lucide-react';

export const SystemHealth: React.FC = () => {
  const [logs, setLogs] = useState<string[]>([
    '[2026-11-03 14:05:00] [INFO] Go Core HTTP Server listening on :8080',
    '[2026-11-03 14:05:00] [INFO] Connected to NATS JetStream on localhost:4222',
    '[2026-11-03 14:05:00] [INFO] Connected to PostgreSQL with pgvector',
    '[2026-11-03 14:05:00] [INFO] Connected to Neo4j graph cluster',
    '[2026-11-03 14:05:01] [INFO] Background worker started listening for PIPELINE_JOBS',
    '[2026-11-03 15:10:22] [INFO] IngestDirectory: chronos-self ready for semantic search & RAG queries',
  ]);

  const services = [
    { name: 'Go Core API', addr: 'http://localhost:8080', status: 'ONLINE' },
    { name: 'Python AI (gRPC)', addr: 'localhost:50052', status: 'ONLINE' },
    { name: 'Rust Parser (gRPC)', addr: 'localhost:50051', status: 'ONLINE' },
    { name: 'PostgreSQL (pgvector)', addr: 'localhost:5432', status: 'ONLINE' },
    { name: 'Neo4j Graph DB', addr: 'bolt://localhost:7687', status: 'ONLINE' },
    { name: 'NATS JetStream', addr: 'nats://localhost:4222', status: 'ONLINE' },
  ];

  return (
    <div className="flex-1 overflow-y-auto p-8 space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-white">System Health & Live Telemetry</h1>
        <p className="text-sm text-slate-400 mt-1">Trạng thái NATS JetStream, gRPC Microservices và luồng logs thời gian thực.</p>
      </div>

      {/* Services Grid */}
      <div className="grid grid-cols-3 gap-4">
        {services.map((s, idx) => (
          <div key={idx} className="bg-[#131b2e] border border-[#1e293b] p-4 rounded-xl flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-2.5 h-2.5 rounded-full bg-emerald-400 shadow-lg shadow-emerald-400/50"></div>
              <div>
                <div className="font-bold text-sm text-white">{s.name}</div>
                <div className="text-xs text-slate-400 font-mono">{s.addr}</div>
              </div>
            </div>
            <span className="text-xs font-mono text-emerald-400 px-2 py-1 rounded bg-emerald-950/60 border border-emerald-800">
              {s.status}
            </span>
          </div>
        ))}
      </div>

      {/* Live Terminal Logs */}
      <div className="bg-black/90 border border-[#1e293b] rounded-xl overflow-hidden flex flex-col h-96">
        <div className="h-10 bg-[#0d1322] border-b border-[#1e293b] px-4 flex items-center justify-between text-xs text-slate-400">
          <div className="flex items-center gap-2">
            <Terminal className="w-4 h-4 text-emerald-400" />
            <span className="font-mono font-semibold text-slate-300">Telemetry Log Stream</span>
          </div>
          <button onClick={() => setLogs([])} className="hover:text-slate-200">
            Clear
          </button>
        </div>
        <div className="flex-1 p-4 overflow-y-auto font-mono text-xs text-emerald-400/90 space-y-1.5 selection:bg-emerald-800">
          {logs.map((log, idx) => (
            <div key={idx}>{log}</div>
          ))}
        </div>
      </div>
    </div>
  );
};
