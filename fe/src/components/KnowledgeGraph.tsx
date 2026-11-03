import React, { useEffect, useRef, useState } from 'react';
import cytoscape from 'cytoscape';
import { Crosshair, Zap, RotateCcw } from 'lucide-react';

interface KnowledgeGraphProps {
  activeRepo: string;
}

export const KnowledgeGraph: React.FC<KnowledgeGraphProps> = () => {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const cyRef = useRef<cytoscape.Core | null>(null);
  const [impactInput, setImpactInput] = useState('AskCmd');
  const [selectedNode, setSelectedNode] = useState<any | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;

    const elements = [
      // Files (Hexagons - Orange)
      { data: { id: 'file_main', label: 'cmd/chronos/main.go', type: 'file' } },
      { data: { id: 'file_ask', label: 'cmd/cli/commands/ask.go', type: 'file' } },
      { data: { id: 'file_query', label: 'internal/core/query/service.go', type: 'file' } },
      { data: { id: 'file_ingest', label: 'internal/core/ingest/service.go', type: 'file' } },
      { data: { id: 'file_ai', label: 'internal/adapters/grpc/ai_client.go', type: 'file' } },

      // Functions (Circles - Blue)
      { data: { id: 'fn_main', label: 'main()', type: 'func', file: 'cmd/chronos/main.go', lines: '14-40' } },
      { data: { id: 'fn_init', label: 'init()', type: 'func', file: 'cmd/cli/commands/ask.go', lines: '37-41' } },
      { data: { id: 'fn_ask', label: 'Ask()', type: 'func', file: 'internal/core/query/service.go', lines: '30-52' } },
      { data: { id: 'fn_ask_stream', label: 'AskStream()', type: 'func', file: 'internal/core/query/service.go', lines: '54-85' } },
      { data: { id: 'fn_process_file', label: 'ProcessFile()', type: 'func', file: 'internal/core/ingest/service.go', lines: '27-100' } },
      { data: { id: 'fn_query_stream', label: 'QueryStream()', type: 'func', file: 'internal/adapters/grpc/ai_client.go', lines: '58-113' } },
      { data: { id: 'fn_extract_code', label: 'extractCodeByLines()', type: 'func', file: 'internal/core/ingest/service.go', lines: '103-110' } },

      // Edges: BELONGS_TO
      { data: { source: 'fn_main', target: 'file_main', label: 'BELONGS_TO' } },
      { data: { source: 'fn_init', target: 'file_ask', label: 'BELONGS_TO' } },
      { data: { source: 'fn_ask', target: 'file_query', label: 'BELONGS_TO' } },
      { data: { source: 'fn_ask_stream', target: 'file_query', label: 'BELONGS_TO' } },
      { data: { source: 'fn_process_file', target: 'file_ingest', label: 'BELONGS_TO' } },
      { data: { source: 'fn_extract_code', target: 'file_ingest', label: 'BELONGS_TO' } },
      { data: { source: 'fn_query_stream', target: 'file_ai', label: 'BELONGS_TO' } },

      // Edges: CALLS
      { data: { source: 'fn_ask', target: 'fn_ask_stream', label: 'CALLS' } },
      { data: { source: 'fn_ask_stream', target: 'fn_query_stream', label: 'CALLS' } },
      { data: { source: 'fn_process_file', target: 'fn_extract_code', label: 'CALLS' } },
    ];

    cyRef.current = cytoscape({
      container: containerRef.current,
      elements: elements,
      style: [
        {
          selector: 'node[type="func"]',
          style: {
            'background-color': '#0ea5e9',
            'label': 'data(label)',
            'color': '#f8fafc',
            'font-size': '11px',
            'font-family': 'monospace',
            'text-valign': 'bottom',
            'text-margin-y': 6,
            'width': 34,
            'height': 34,
            'border-width': 2,
            'border-color': '#38bdf8',
          },
        },
        {
          selector: 'node[type="file"]',
          style: {
            'background-color': '#f59e0b',
            'shape': 'hexagon',
            'label': 'data(label)',
            'color': '#94a3b8',
            'font-size': '10px',
            'font-family': 'monospace',
            'text-valign': 'bottom',
            'text-margin-y': 6,
            'width': 28,
            'height': 28,
          },
        },
        {
          selector: 'edge',
          style: {
            'width': 1.5,
            'line-color': '#334155',
            'target-arrow-color': '#334155',
            'target-arrow-shape': 'triangle',
            'curve-style': 'bezier',
            'arrow-scale': 0.8,
          },
        },
        {
          selector: '.impact-target',
          style: {
            'background-color': '#ef4444 !important',
            'border-color': '#fca5a5 !important',
            'border-width': 6,
            'width': 44,
            'height': 44,
          },
        },
        {
          selector: '.impact-caller',
          style: {
            'background-color': '#facc15 !important',
            'border-color': '#fef08a !important',
            'border-width': 4,
          },
        },
        {
          selector: '.dimmed',
          style: {
            'opacity': 0.15,
          },
        },
      ],
      layout: {
        name: 'cose',
        padding: 50,
      },
    });

    cyRef.current.on('tap', 'node', (evt) => {
      setSelectedNode(evt.target.data());
    });

    return () => {
      cyRef.current?.destroy();
      cyRef.current = null;
    };
  }, []);

  const handleImpact = async () => {
    if (!cyRef.current || !impactInput.trim()) return;

    try {
      await fetch(`/api/v1/impact?function=${encodeURIComponent(impactInput.trim())}`);
    } catch (e) {}

    cyRef.current.elements().removeClass('impact-target impact-caller dimmed');
    cyRef.current.elements().addClass('dimmed');

    const target = cyRef.current.nodes().filter((n) => n.data('label').toLowerCase().includes(impactInput.toLowerCase()));
    target.removeClass('dimmed').addClass('impact-target');

    const callers = target.incomers('edge').sources();
    callers.removeClass('dimmed').addClass('impact-caller');
    target.incomers('edge').removeClass('dimmed');
  };

  const handleReset = () => {
    if (!cyRef.current) return;
    cyRef.current.elements().removeClass('impact-target impact-caller dimmed');
    setSelectedNode(null);
  };

  return (
    <div className="flex-1 flex flex-col h-full relative overflow-hidden">
      {/* Top floating bar */}
      <div className="absolute top-4 left-4 right-4 z-10 flex items-center justify-between gap-4 pointer-events-none">
        <div className="bg-[#0d1322]/90 backdrop-blur border border-[#1e293b] p-2 rounded-xl shadow-xl flex items-center gap-2 pointer-events-auto">
          <div className="flex items-center gap-2 px-2 text-xs font-semibold text-rose-400 uppercase tracking-wider">
            <Crosshair className="w-4 h-4" /> Impact Tool:
          </div>
          <input
            type="text"
            value={impactInput}
            onChange={(e) => setImpactInput(e.target.value)}
            placeholder="Tên hàm (vd: AskCmd, ProcessFile)..."
            className="bg-[#0b1120] border border-[#1e293b] text-xs px-3 py-1.5 rounded-lg text-slate-200 outline-none focus:border-rose-500 w-48"
          />
          <button
            onClick={handleImpact}
            className="bg-rose-500 hover:bg-rose-600 text-white font-semibold text-xs px-3 py-1.5 rounded-lg transition flex items-center gap-1.5"
          >
            <Zap className="w-3.5 h-3.5" /> Blast Radius
          </button>
          <button onClick={handleReset} className="text-xs text-slate-400 hover:text-slate-200 px-2 py-1 flex items-center gap-1">
            <RotateCcw className="w-3 h-3" /> Reset
          </button>
        </div>

        <div className="bg-[#0d1322]/90 backdrop-blur border border-[#1e293b] px-3 py-2 rounded-xl text-xs flex items-center gap-4 pointer-events-auto text-slate-300">
          <span className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded-full bg-sky-500"></span> Function
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded bg-amber-500"></span> File
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded-full bg-rose-500"></span> Target Impact
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded-full bg-yellow-400"></span> Callers (Affected)
          </span>
        </div>
      </div>

      {/* Canvas */}
      <div ref={containerRef} className="flex-1 w-full h-full bg-[#090d16]"></div>

      {/* Node Inspector Popup */}
      {selectedNode && (
        <div className="absolute bottom-4 left-4 right-4 bg-[#0d1322]/95 backdrop-blur border border-[#1e293b] p-4 rounded-xl shadow-2xl flex items-center justify-between text-xs z-10">
          <div className="space-y-1">
            <div className="font-bold text-sm text-white flex items-center gap-2">
              <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-sky-950 text-sky-400 border border-sky-800 uppercase">
                {selectedNode.type}
              </span>
              <span>{selectedNode.label}</span>
            </div>
            <div className="text-slate-400 font-mono">
              {selectedNode.file || 'File'} (Lines: {selectedNode.lines || 'N/A'})
            </div>
          </div>
          <button
            onClick={() => {
              setImpactInput(selectedNode.label.replace('()', ''));
              handleImpact();
            }}
            className="bg-rose-500/20 text-rose-400 border border-rose-500/40 px-3 py-1.5 rounded-lg hover:bg-rose-500/30 transition"
          >
            Check Blast Radius
          </button>
        </div>
      )}
    </div>
  );
};
