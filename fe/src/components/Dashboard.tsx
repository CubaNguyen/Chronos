import React, { useState } from 'react';
import { Files, Code2, Database, Share2, Send, FolderPlus, FolderGit2, ArrowRight, Check, X } from 'lucide-react';

interface DashboardProps {
  onNavigateToChat: () => void;
}

export const Dashboard: React.FC<DashboardProps> = ({ onNavigateToChat }) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [repoName, setRepoName] = useState('chronos-self');
  const [dirPath, setDirPath] = useState('c:\\Users\\huygn\\OneDrive\\Desktop\\ptit hihi nho push\\Chronos');
  const [isIngesting, setIsIngesting] = useState(false);
  const [progress, setProgress] = useState(0);
  const [statusText, setStatusText] = useState('');

  const handleStartIngest = async () => {
    setIsIngesting(true);
    setProgress(20);
    setStatusText('Publishing jobs to NATS queue (/api/v1/ingest-dir)...');

    try {
      await fetch('/api/v1/ingest-dir', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ repo_name: repoName, dir_path: dirPath }),
      });

      setTimeout(() => {
        setProgress(60);
        setStatusText('Rust Parser extracting AST & Python generating embeddings...');
      }, 1200);

      setTimeout(() => {
        setProgress(100);
        setStatusText('Ingestion completed! Knowledge Base updated.');
        setTimeout(() => {
          setIsModalOpen(false);
          setIsIngesting(false);
          setProgress(0);
        }, 1500);
      }, 3000);
    } catch (err: any) {
      setStatusText(`Error: ${err.message}`);
      setIsIngesting(false);
    }
  };

  return (
    <div className="flex-1 overflow-y-auto p-8 space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white">Repository Overview</h1>
          <p className="text-sm text-slate-400 mt-1">Quản lý các kho mã nguồn và số liệu tri thức trong Knowledge Base.</p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 bg-gradient-to-r from-sky-500 to-indigo-600 hover:from-sky-400 hover:to-indigo-500 text-white font-medium px-4 py-2.5 rounded-lg shadow-lg shadow-sky-500/20 text-sm transition"
        >
          <FolderPlus className="w-4 h-4" />
          <span>Add Repository</span>
        </button>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-5 gap-4">
        <div className="bg-[#131b2e] border border-[#1e293b] rounded-xl p-5 relative overflow-hidden">
          <div className="flex items-center justify-between text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">
            <span>Total Files</span>
            <Files className="w-4 h-4 text-sky-400" />
          </div>
          <div className="text-2xl font-extrabold text-white">38</div>
          <div className="text-[11px] text-emerald-400 mt-1 flex items-center gap-1">
            Parsed by Rust <Check className="w-3 h-3" />
          </div>
        </div>

        <div className="bg-[#131b2e] border border-[#1e293b] rounded-xl p-5 relative overflow-hidden">
          <div className="flex items-center justify-between text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">
            <span>Functions</span>
            <Code2 className="w-4 h-4 text-indigo-400" />
          </div>
          <div className="text-2xl font-extrabold text-white">128</div>
          <div className="text-[11px] text-slate-400 mt-1">Chunked AST nodes</div>
        </div>

        <div className="bg-[#131b2e] border border-[#1e293b] rounded-xl p-5 relative overflow-hidden">
          <div className="flex items-center justify-between text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">
            <span>Vectors (Postgres)</span>
            <Database className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-extrabold text-white">128</div>
          <div className="text-[11px] text-emerald-400 mt-1">384d all-MiniLM-L6-v2</div>
        </div>

        <div className="bg-[#131b2e] border border-[#1e293b] rounded-xl p-5 relative overflow-hidden">
          <div className="flex items-center justify-between text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">
            <span>Graph Nodes (Neo4j)</span>
            <Share2 className="w-4 h-4 text-amber-400" />
          </div>
          <div className="text-2xl font-extrabold text-white">216</div>
          <div className="text-[11px] text-amber-400 mt-1">Functions & Files mapped</div>
        </div>

        <div className="bg-[#131b2e] border border-[#1e293b] rounded-xl p-5 relative overflow-hidden">
          <div className="flex items-center justify-between text-slate-400 text-xs font-semibold uppercase tracking-wider mb-2">
            <span>NATS Jobs</span>
            <Send className="w-4 h-4 text-purple-400" />
          </div>
          <div className="text-2xl font-extrabold text-white">100%</div>
          <div className="text-[11px] text-purple-400 mt-1">Queue processed</div>
        </div>
      </div>

      {/* Repo Cards */}
      <div>
        <h2 className="text-sm font-semibold uppercase tracking-wider text-slate-400 mb-4">Indexed Projects</h2>
        <div className="grid grid-cols-3 gap-5">
          <div
            onClick={onNavigateToChat}
            className="bg-[#131b2e] border border-sky-500/30 rounded-xl p-5 hover:border-sky-500/60 transition group cursor-pointer"
          >
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <FolderGit2 className="w-5 h-5 text-sky-400" />
                <h3 className="font-bold text-white text-base">chronos-self</h3>
              </div>
              <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-950 text-emerald-400 border border-emerald-800">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-400"></span> Indexed
              </span>
            </div>
            <p className="text-xs text-slate-400 mb-4 line-clamp-2">
              Toàn bộ mã nguồn hệ thống Chronos (Go Core, Rust Parser, Python AI, gRPC & Web UI).
            </p>
            <div className="flex items-center justify-between pt-3 border-t border-[#1e293b] text-xs text-slate-400">
              <span>38 Files • 128 Chunks</span>
              <span className="text-sky-400 group-hover:translate-x-1 transition flex items-center gap-1 font-medium">
                Ask AI <ArrowRight className="w-3.5 h-3.5" />
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Ingest Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 bg-black/70 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-[#131b2e] border border-[#1e293b] rounded-2xl w-full max-w-md p-6 space-y-5 shadow-2xl">
            <div className="flex items-center justify-between">
              <h3 className="font-bold text-lg text-white flex items-center gap-2">
                <FolderPlus className="w-5 h-5 text-sky-400" /> Ingest New Repository
              </h3>
              <button onClick={() => setIsModalOpen(false)} className="text-slate-400 hover:text-white">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">Repository Name</label>
                <input
                  type="text"
                  value={repoName}
                  onChange={(e) => setRepoName(e.target.value)}
                  className="w-full bg-[#0b1120] border border-[#1e293b] text-sm text-slate-200 px-3.5 py-2.5 rounded-lg focus:outline-none focus:border-sky-500"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">Directory Path</label>
                <input
                  type="text"
                  value={dirPath}
                  onChange={(e) => setDirPath(e.target.value)}
                  className="w-full bg-[#0b1120] border border-[#1e293b] text-sm text-slate-200 px-3.5 py-2.5 rounded-lg focus:outline-none focus:border-sky-500 font-mono text-xs"
                />
              </div>
            </div>

            {isIngesting && (
              <div className="space-y-2">
                <div className="flex justify-between text-xs text-slate-400">
                  <span>{statusText}</span>
                  <span>{progress}%</span>
                </div>
                <div className="w-full bg-[#0b1120] h-2 rounded-full overflow-hidden border border-[#1e293b]">
                  <div className="bg-sky-500 h-full transition-all duration-300" style={{ width: `${progress}%` }}></div>
                </div>
              </div>
            )}

            <div className="flex justify-end gap-3 pt-2">
              <button onClick={() => setIsModalOpen(false)} className="px-4 py-2 rounded-lg text-sm text-slate-400 hover:text-white transition">
                Cancel
              </button>
              <button
                disabled={isIngesting}
                onClick={handleStartIngest}
                className="bg-sky-500 hover:bg-sky-400 text-slate-900 font-semibold px-4 py-2 rounded-lg text-sm transition disabled:opacity-50"
              >
                Start Ingestion
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
