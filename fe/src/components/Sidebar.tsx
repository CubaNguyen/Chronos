import React from 'react';
import { LayoutDashboard, MessageSquareCode, Search, Network, Activity, Cpu, ChevronDown, RefreshCw } from 'lucide-react';

interface SidebarProps {
  activeTab: string;
  setActiveTab: (tab: string) => void;
  activeRepo: string;
  setActiveRepo: (repo: string) => void;
}

export const Sidebar: React.FC<SidebarProps> = ({ activeTab, setActiveTab, activeRepo, setActiveRepo }) => {
  const menu = [
    { id: 'dashboard', label: '1. Dashboard', icon: LayoutDashboard },
    { id: 'chat', label: '2. AI Chat (RAG)', icon: MessageSquareCode },
    { id: 'search', label: '3. Semantic Search', icon: Search },
    { id: 'graph', label: '4. Knowledge Graph', icon: Network },
    { id: 'health', label: '5. System Health', icon: Activity },
  ];

  return (
    <aside className="w-64 bg-[#0d1322] border-r border-[#1e293b] flex flex-col justify-between flex-shrink-0 select-none">
      <div>
        {/* Brand */}
        <div className="h-16 flex items-center px-6 border-b border-[#1e293b] gap-3">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-tr from-sky-500 to-indigo-500 flex items-center justify-center shadow-lg shadow-sky-500/20">
            <Cpu className="w-5 h-5 text-white" />
          </div>
          <div>
            <span className="font-bold text-lg bg-gradient-to-r from-sky-400 to-indigo-300 bg-clip-text text-transparent">Chronos</span>
            <span className="text-[10px] text-sky-400 font-mono ml-1.5 px-1.5 py-0.5 rounded bg-sky-950 border border-sky-800">React</span>
          </div>
        </div>

        {/* Menu Items */}
        <nav className="p-3 space-y-1">
          {menu.map((item) => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => setActiveTab(item.id)}
                className={`w-full flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition ${
                  isActive
                    ? 'text-sky-400 bg-sky-950/50 border border-sky-800/60 shadow-sm'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-[#131b2e]/50'
                }`}
              >
                <Icon className="w-4 h-4" />
                <span>{item.label}</span>
              </button>
            );
          })}
        </nav>
      </div>

      {/* Target Repo */}
      <div className="p-4 border-t border-[#1e293b] space-y-3 bg-[#090d16]/40">
        <div>
          <label className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block mb-1">Target Repository</label>
          <div className="relative">
            <select
              value={activeRepo}
              onChange={(e) => setActiveRepo(e.target.value)}
              className="w-full bg-[#131b2e] border border-[#1e293b] text-xs text-slate-200 rounded-lg px-2.5 py-2 appearance-none focus:outline-none focus:border-sky-500"
            >
              <option value="chronos-self">chronos-self (Indexed)</option>
            </select>
            <ChevronDown className="w-3.5 h-3.5 text-slate-400 absolute right-2.5 top-2.5 pointer-events-none" />
          </div>
        </div>
        <div className="flex items-center justify-between text-xs text-slate-400 pt-1">
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
            Core API :8080
          </span>
          <RefreshCw className="w-3.5 h-3.5 hover:text-sky-400 cursor-pointer" />
        </div>
      </div>
    </aside>
  );
};

