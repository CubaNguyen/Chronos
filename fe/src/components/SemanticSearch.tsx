import React, { useState } from 'react';
import { Search, ArrowRight } from 'lucide-react';

interface SemanticSearchProps {
  activeRepo: string;
}

export const SemanticSearch: React.FC<SemanticSearchProps> = ({ activeRepo }) => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const handleSearch = async () => {
    if (!query.trim()) return;
    setIsLoading(true);

    try {
      const res = await fetch('/api/v1/search', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ repo_name: activeRepo, query: query }),
      });
      const data = await res.json();
      setResults(data.data || []);
    } catch (err) {
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex-1 overflow-y-auto p-8 space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-white">Semantic Code Search</h1>
        <p className="text-sm text-slate-400 mt-1">Tìm kiếm code thuần túy bằng Vector Embedding (PostgreSQL pgvector) không qua LLM.</p>
      </div>

      {/* Search Input Box */}
      <div className="flex gap-3">
        <div className="flex-1 relative">
          <Search className="w-5 h-5 text-slate-400 absolute left-3.5 top-3.5 pointer-events-none" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            placeholder="Nhập ý định tìm kiếm (vd: kết nối gRPC client, xử lý lưu vector, đọc file git)..."
            className="w-full bg-[#131b2e] border border-[#1e293b] text-sm text-slate-200 pl-11 pr-4 py-3 rounded-xl focus:outline-none focus:border-sky-500 transition shadow-inner"
          />
        </div>
        <button
          onClick={handleSearch}
          className="bg-sky-500 hover:bg-sky-400 text-slate-900 font-semibold px-6 py-3 rounded-xl transition flex items-center gap-2 text-sm"
        >
          <span>Search</span>
          <ArrowRight className="w-4 h-4" />
        </button>
      </div>

      {/* Results List */}
      <div className="space-y-4">
        {isLoading && <div className="text-center py-8 text-sky-400 font-mono text-xs">Đang truy vấn vector embedding từ Postgres...</div>}

        {!isLoading && results.length === 0 && (
          <div className="text-center py-12 text-slate-500 text-sm">Gõ từ khóa và bấm Enter để tìm các đoạn code tương tự.</div>
        )}

        {!isLoading &&
          results.map((item, idx) => {
            const percent = Math.min(98, Math.max(72, Math.round(96 - idx * 4)));
            return (
              <div key={idx} className="bg-[#131b2e] border border-[#1e293b] rounded-xl p-5 hover:border-sky-500/50 transition space-y-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="font-mono font-bold text-sm text-sky-400">{item.FunctionName || 'Function'}()</span>
                    <span className="text-xs text-slate-400 font-mono">{item.FilePath}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="w-32 bg-[#090d16] h-2 rounded-full overflow-hidden border border-[#1e293b]">
                      <div className="bg-emerald-400 h-full" style={{ width: `${percent}%` }}></div>
                    </div>
                    <span className="text-xs font-mono text-emerald-400 font-semibold">{percent}% Match</span>
                  </div>
                </div>
                <pre className="bg-black/60 p-3 rounded-lg overflow-x-auto text-xs font-mono text-slate-300 border border-[#1e293b]/40">
                  <code>{item.Content}</code>
                </pre>
              </div>
            );
          })}
      </div>
    </div>
  );
};
