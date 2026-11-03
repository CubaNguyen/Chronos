import React, { useState, useRef } from 'react';
import { Bot, User, Send, Square, FileCode } from 'lucide-react';

interface AIChatProps {
  activeRepo: string;
}

interface Message {
  role: 'user' | 'ai';
  text: string;
  citations?: { file: string; start: number; end: number }[];
}

const defaultCodeSnippets: Record<string, { title: string; lang: string; lines: string[] }> = {
  'ask.go': {
    title: 'cmd/cli/commands/ask.go',
    lang: 'go',
    lines: [
      'package commands',
      '',
      'import (',
      '\t"context"',
      '\t"fmt"',
      '\t"log"',
      '\t"github.com/cuba3/chronos/core/internal/app"',
      '\t"github.com/cuba3/chronos/core/internal/config"',
      '\t"github.com/spf13/cobra"',
      ')',
      '',
      'var (',
      '\taskRepo     string',
      '\taskQuestion string',
      ')',
      '',
      'var AskCmd = &cobra.Command{',
      '\tUse:   "ask",',
      '\tShort: "Đặt câu hỏi về mã nguồn",',
      '\tRun: func(cmd *cobra.Command, args []string) {',
      '\t\tif askQuestion == "" {',
      '\t\t\tfmt.Println("Vui lòng nhập câu hỏi bằng cờ --question hoặc -q")',
      '\t\t\treturn',
      '\t\t}',
      '',
      '\t\tcfg, _ := config.Load()',
      '\t\tcoreApp, err := app.NewCoreApp(cfg)',
      '\t\tif err != nil { log.Fatalf("Lỗi: %v", err) }',
      '\t\tdefer coreApp.Close()',
      '',
      '\t\tresult, err := coreApp.QueryService.Ask(context.Background(), askRepo, askQuestion)',
      '\t\tif err != nil { log.Fatalf("Lỗi truy vấn: %v", err) }',
      '\t\tfmt.Println(result.Answer)',
      '\t},',
      '}',
      '',
      'func init() {',
      '\tAskCmd.Flags().StringVarP(&askRepo, "repo", "r", "chronos-self", "Tên dự án")',
      '\tAskCmd.Flags().StringVarP(&askQuestion, "question", "q", "", "Câu hỏi của bạn")',
      '\tAskCmd.MarkFlagRequired("question")',
      '}',
    ],
  },
  'git_service.go': {
    title: 'internal/core/ingest/git_service.go',
    lang: 'go',
    lines: [
      'package ingest',
      '',
      'func (s *GitIngestService) IngestDirectory(ctx context.Context, repoName string, dirPath string) error {',
      '\tslog.Info("Quét thư mục", "path", dirPath)',
      '\treturn filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {',
      '\t\tif info.IsDir() || !isSupported(path) { return nil }',
      '\t\tsource, _ := os.ReadFile(path)',
      '\t\treturn s.queue.Publish(ctx, repoName, path, string(source))',
      '\t})',
      '}',
    ],
  },
};

export const AIChat: React.FC<AIChatProps> = ({ activeRepo }) => {
  const [messages, setMessages] = useState<Message[]>([
    {
      role: 'ai',
      text: `Tôi đã lập chỉ mục dự án \`${activeRepo}\`. Hãy đặt câu hỏi về mã nguồn, tôi sẽ stream câu trả lời và đính kèm trích dẫn source code.`,
    },
  ]);
  const [inputQuestion, setInputQuestion] = useState('');
  const [isStreaming, setIsStreaming] = useState(false);
  const [selectedFile, setSelectedFile] = useState('ask.go');
  const [highlightRange, setHighlightRange] = useState<[number, number] | null>([37, 41]);

  const abortControllerRef = useRef<AbortController | null>(null);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSend = async (q?: string) => {
    const question = (q || inputQuestion).trim();
    if (!question || isStreaming) return;

    setInputQuestion('');
    setMessages((prev) => [...prev, { role: 'user', text: question }, { role: 'ai', text: '' }]);
    setIsStreaming(true);

    abortControllerRef.current = new AbortController();
    let accumulated = '';

    try {
      const res = await fetch('/api/v1/ask-stream', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ repo_name: activeRepo, question: question }),
        signal: abortControllerRef.current.signal,
      });

      const reader = res.body?.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      if (reader) {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split('\n');
          buffer = lines.pop() || '';

          for (const line of lines) {
            const trimmed = line.trim();
            if (trimmed.startsWith('data:')) {
              try {
                const parsed = JSON.parse(trimmed.slice(5).trim() || '{}');
                if (parsed.answer) {
                  accumulated += parsed.answer;
                  setMessages((prev) => {
                    const copy = [...prev];
                    copy[copy.length - 1] = {
                      role: 'ai',
                      text: accumulated,
                      citations: [
                        { file: 'ask.go', start: 37, end: 41 },
                        { file: 'git_service.go', start: 3, end: 8 },
                      ],
                    };
                    return copy;
                  });
                  scrollToBottom();
                }
              } catch (e) {}
            }
          }
        }
      }
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        setMessages((prev) => {
          const copy = [...prev];
          copy[copy.length - 1] = { role: 'ai', text: accumulated + `\n\n❌ Lỗi: ${err.message}` };
          return copy;
        });
      }
    } finally {
      setIsStreaming(false);
      abortControllerRef.current = null;
    }
  };

  const handleStop = () => {
    abortControllerRef.current?.abort();
  };

  const codeData = defaultCodeSnippets[selectedFile] || defaultCodeSnippets['ask.go'];

  return (
    <div className="flex-1 flex flex-row h-full overflow-hidden">
      {/* Left Panel: Chat Window */}
      <div className="w-1/2 flex flex-col h-full border-r border-[#1e293b] bg-[#090d16]/60">
        <div className="flex-1 overflow-y-auto p-5 space-y-4">
          {messages.map((m, idx) => (
            <div
              key={idx}
              className={`flex flex-col gap-2 max-w-[90%] p-4 rounded-xl text-sm leading-relaxed ${
                m.role === 'user' ? 'bg-sky-600 text-white self-end ml-auto' : 'bg-[#131b2e] border border-[#1e293b] text-slate-200'
              }`}
            >
              <div
                className={`flex items-center gap-2 text-xs font-semibold uppercase tracking-wider ${
                  m.role === 'user' ? 'text-sky-200' : 'text-sky-400'
                }`}
              >
                {m.role === 'user' ? <User className="w-3.5 h-3.5" /> : <Bot className="w-4 h-4" />}
                <span>{m.role === 'user' ? 'Bạn' : 'Chronos Assistant'}</span>
              </div>
              <div className="whitespace-pre-wrap">{m.text || <span className="italic text-slate-500">Đang suy nghĩ...</span>}</div>

              {m.citations && (
                <div className="pt-2 border-t border-[#1e293b]/60 space-y-1">
                  <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Citations:</div>
                  <div className="flex gap-2 flex-wrap">
                    {m.citations.map((c, cIdx) => (
                      <button
                        key={cIdx}
                        onClick={() => {
                          setSelectedFile(c.file);
                          setHighlightRange([c.start, c.end]);
                        }}
                        className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded bg-sky-950/70 border border-sky-800 text-sky-400 text-xs font-mono hover:bg-sky-900 transition"
                      >
                        <span>[{cIdx + 1}]</span> {c.file}:{c.start}
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ))}
          <div ref={messagesEndRef} />
        </div>

        {/* Suggestion Chips */}
        <div className="px-5 py-2 flex flex-wrap gap-2 border-t border-[#1e293b]/50 bg-[#0d1322]/40">
          <button
            onClick={() => handleSend('Hàm init trong file ask.go làm gì?')}
            className="text-xs bg-[#131b2e] hover:bg-slate-800 text-slate-300 border border-[#1e293b] px-2.5 py-1.5 rounded-lg transition"
          >
            💡 init() trong ask.go
          </button>
          <button
            onClick={() => handleSend('Hàm IngestDirectory trong git_service.go hoạt động như thế nào?')}
            className="text-xs bg-[#131b2e] hover:bg-slate-800 text-slate-300 border border-[#1e293b] px-2.5 py-1.5 rounded-lg transition"
          >
            💡 IngestDirectory()
          </button>
        </div>

        {/* Input Footer */}
        <div className="p-4 bg-[#0d1322] border-t border-[#1e293b]">
          <div className="flex items-end gap-2 bg-[#0b1120] border border-[#1e293b] rounded-xl p-2.5 focus-within:border-sky-500 transition">
            <textarea
              rows={1}
              value={inputQuestion}
              onChange={(e) => setInputQuestion(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault();
                  handleSend();
                }
              }}
              placeholder="Hỏi về mã nguồn... (Enter để gửi, Shift+Enter xuống dòng)"
              className="flex-1 bg-transparent text-sm text-slate-200 placeholder-slate-500 resize-none outline-none max-h-32 leading-relaxed"
            />
            {isStreaming ? (
              <button onClick={handleStop} className="bg-rose-500 hover:bg-rose-400 text-white font-semibold p-2 rounded-lg transition">
                <Square className="w-4 h-4" />
              </button>
            ) : (
              <button onClick={() => handleSend()} className="bg-sky-500 hover:bg-sky-400 text-slate-900 font-semibold p-2 rounded-lg transition">
                <Send className="w-4 h-4" />
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Right Panel: Code Viewer */}
      <div className="w-1/2 flex flex-col h-full bg-[#090d16]">
        <div className="h-12 border-b border-[#1e293b] px-4 flex items-center justify-between bg-[#0d1322]">
          <div className="flex items-center gap-2">
            <FileCode className="w-4 h-4 text-sky-400" />
            <span className="font-mono text-xs font-semibold text-slate-300">{codeData.title}</span>
          </div>
          <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-400 uppercase">{codeData.lang}</span>
        </div>
        <div className="flex-1 overflow-auto font-mono text-xs p-4 bg-[#0d1117] text-slate-300">
          {codeData.lines.map((line, idx) => {
            const lineNum = idx + 1;
            const isHighlighted = highlightRange && lineNum >= highlightRange[0] && lineNum <= highlightRange[1];
            return (
              <div key={idx} className={`code-line ${isHighlighted ? 'highlighted' : ''}`}>
                <span className="line-number">{lineNum}</span>
                <span className="line-content">{line}</span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

