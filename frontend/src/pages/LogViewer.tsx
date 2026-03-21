import { useEffect, useRef, useState } from "react";
import {
  GetRecentLogs,
  StartLogWatch,
  StopLogWatch,
} from "../../wailsjs/go/main/App";
import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime";
import { ServiceIcon } from "../components/ServiceIcon";
import { useI18n, tt } from '../i18n';
import { FileText, Play, Pause, ArrowDown } from 'lucide-react';

interface LogEntry {
  service: string;
  line: string;
  level: "error" | "warning" | "info" | "debug";
  timestamp: string;
}

const SERVICES = [
  { id: 'apache', label: 'Apache' },
  { id: 'nginx',  label: 'Nginx'  },
  { id: 'mysql',  label: 'MySQL' },
  { id: 'php',    label: 'PHP'},
]

const levelStyles: Record<string, string> = {
  error: "text-red-400",
  warning: "text-yellow-400",
  info: "text-gray-300",
  debug: "text-gray-500",
};

const levelBadge: Record<string, string> = {
  error: "bg-red-500/20 text-red-400",
  warning: "bg-yellow-500/20 text-yellow-400",
  info: "bg-blue-500/20 text-blue-400",
  debug: "bg-gray-500/20 text-gray-400",
};

export default function LogViewer() {
  const { t } = useI18n();
  const [activeService, setActiveService] = useState("apache");
  const [entries, setEntries] = useState<LogEntry[]>([]);
  const [filter, setFilter] = useState("");
  const [levelFilter, setLevelFilter] = useState<string>("all");
  const [autoScroll, setAutoScroll] = useState(true);
  const [paused, setPaused] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);
  const pausedRef = useRef(false);

  pausedRef.current = paused;

  // Load recent logs + start watching khi đổi service
  useEffect(() => {
    setEntries([]);

    // Load 200 dòng gần nhất
    GetRecentLogs(activeService, 200).then((data) => {
      setEntries((data || []) as LogEntry[]);
    });

    // Bắt đầu watch realtime
    StartLogWatch(activeService);

    return () => {
      StopLogWatch();
    };
  }, [activeService]);

  // Lắng nghe log mới từ backend
  useEffect(() => {
    const handler = (entry: LogEntry) => {
      if (pausedRef.current) return;
      setEntries((prev) => {
        const next = [...prev, entry];
        // Giới hạn 1000 dòng trong memory
        return next.length > 1000 ? next.slice(-1000) : next;
      });
    };

    EventsOn("log:line", handler as any);
    return () => {
      EventsOff("log:line");
    };
  }, []);

  // Auto scroll xuống dưới khi có log mới
  useEffect(() => {
    if (autoScroll && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [entries, autoScroll]);

  const filteredEntries = entries.filter((e) => {
    const matchText =
      filter === "" || e.line.toLowerCase().includes(filter.toLowerCase());
    const matchLevel = levelFilter === "all" || e.level === levelFilter;
    return matchText && matchLevel;
  });

  const errorCount = entries.filter((e) => e.level === "error").length;
  const warningCount = entries.filter((e) => e.level === "warning").length;

  return (
    <div className="flex flex-col h-full gap-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">{t.log_title}</h2>
          <p className="text-gray-400 text-sm mt-1">
            {tt(t.log_lines, { count: entries.length })}
            {errorCount > 0 && (
              <span className="ml-2 text-red-400">{tt(t.log_errors, { count: errorCount })}</span>
            )}
            {warningCount > 0 && (
              <span className="ml-2 text-yellow-400">
                {tt(t.log_warnings, { count: warningCount })}
              </span>
            )}
          </p>
        </div>

        <div className="flex gap-2">
          <button
            onClick={() => setEntries([])}
            className="px-3 py-1.5 text-xs rounded-lg bg-[#1e2535] text-gray-400 hover:text-white transition-colors"
          >
            {t.log_clear}
          </button>
          <button
            onClick={() => setPaused((p) => !p)}
            className={`px-3 py-1.5 text-xs rounded-lg transition-colors flex items-center gap-1 ${
              paused
                ? "bg-yellow-500/20 text-yellow-400 hover:bg-yellow-500/30"
                : "bg-[#1e2535] text-gray-400 hover:text-white"
            }`}
          >
            {paused ? <><Play size={14} /> {t.log_resume}</> : <><Pause size={14} /> {t.log_pause}</>}
          </button>
          <button
            onClick={() => setAutoScroll((a) => !a)}
            className={`px-3 py-1.5 text-xs rounded-lg transition-colors flex items-center gap-1 ${
              autoScroll
                ? "bg-blue-500/20 text-blue-400"
                : "bg-[#1e2535] text-gray-400"
            }`}
          >
            <ArrowDown size={14} /> {t.log_autoscroll}
          </button>
        </div>
      </div>

      {/* Service tabs */}
      <div className="flex gap-1">
        {SERVICES.map((svc) => (
          <button
            key={svc.id}
            onClick={() => setActiveService(svc.id)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors
                      ${
                        activeService === svc.id
                          ? "bg-blue-500/20 text-blue-400 border border-blue-500/30"
                          : "bg-[#1e2535] text-gray-400 hover:text-white border border-transparent"
                      }`}
          >
            <span>
              <ServiceIcon name={svc.id} />
            </span>
            <span>{svc.label}</span>
          </button>
        ))}
      </div>

      {/* Filters */}
      <div className="flex gap-3">
        <input
          type="text"
          placeholder={t.log_filter}
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="flex-1 bg-[#1e2535] border border-[#2a3347] rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
        />
        <select
          value={levelFilter}
          onChange={(e) => setLevelFilter(e.target.value)}
          className="bg-[#1e2535] border border-[#2a3347] rounded-lg px-3 py-2 text-sm text-gray-300 focus:outline-none focus:border-blue-500"
        >
          <option value="all">{t.log_all_levels}</option>
          <option value="error">{t.log_error_level}</option>
          <option value="warning">{t.log_warn_level}</option>
          <option value="info">{t.log_info_level}</option>
          <option value="debug">{t.log_debug_level}</option>
        </select>
      </div>

      {/* Log output */}
      <div className="flex-1 bg-[#080d15] border border-[#1e2535] rounded-xl overflow-auto font-mono text-xs min-h-0">
        {filteredEntries.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-gray-600 gap-2">
            <FileText size={40} />
            <p>{t.log_no_entries}</p>
            <p className="text-gray-700 text-xs">
              {tt(t.log_start_to_see, { service: activeService })}
            </p>
          </div>
        ) : (
          <table className="w-full">
            <tbody>
              {filteredEntries.map((entry, i) => (
                <tr
                  key={i}
                  className="border-b border-[#0f1420] hover:bg-[#0f1420] transition-colors"
                >
                  {/* Timestamp */}
                  <td className="px-3 py-1 text-gray-600 whitespace-nowrap w-16 select-none">
                    {entry.timestamp}
                  </td>

                  {/* Level badge */}
                  <td className="px-2 py-1 w-20">
                    <span
                      className={`px-1.5 py-0.5 rounded text-xs ${levelBadge[entry.level]}`}
                    >
                      {entry.level}
                    </span>
                  </td>

                  {/* Log line */}
                  <td
                    className={`px-3 py-1 break-all ${levelStyles[entry.level]}`}
                  >
                    {highlightFilter(entry.line, filter)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
        <div ref={bottomRef} />
      </div>

      {/* Status bar */}
      <div className="flex items-center justify-between text-xs text-gray-600">
        <span>
          {t.log_watching} <span className="text-gray-400">{activeService}</span>
        </span>
        <span>
          {tt(t.log_lines_shown, { shown: filteredEntries.length, total: entries.length })}
        </span>
      </div>
    </div>
  );
}

// Highlight text khớp với filter
function highlightFilter(line: string, filter: string) {
  if (!filter) return <>{line}</>;
  const idx = line.toLowerCase().indexOf(filter.toLowerCase());
  if (idx === -1) return <>{line}</>;
  return (
    <>
      {line.slice(0, idx)}
      <mark className="bg-yellow-400/30 text-yellow-300 rounded px-0.5">
        {line.slice(idx, idx + filter.length)}
      </mark>
      {line.slice(idx + filter.length)}
    </>
  );
}
