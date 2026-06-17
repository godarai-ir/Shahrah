'use client';

import { motion } from 'motion/react';
import Link from 'next/link';
import { MagicCard } from '../../components/ui/MagicCard';
import { FAQAccordion } from '../../components/ui/FAQAccordion';
import {
  Server, ShieldCheck, Key, Eye, BarChart3, Lock, Network,
  Users, CheckCircle, XCircle, AlertTriangle, ArrowLeft,
  Layers, CreditCard, Shield, Globe, Activity, Zap,
  Database, GitBranch, Building, Package, Wrench, Terminal,
  ChevronRight, FileCode, Settings, BookOpen, Search,
  TrendingDown, Bot, Star, LayoutGrid
} from 'lucide-react';
import { useState, useEffect } from 'react';

// ─── Hero Dashboard ───────────────────────────────────────────────────────────
function MCPGatewayDashboard() {
  const [invocations, setInvocations] = useState(1240);
  const [blocked, setBlocked] = useState(32);

  useEffect(() => {
    const t1 = setInterval(() => setInvocations(v => v + Math.floor(Math.random() * 3 + 1)), 1800);
    const t2 = setInterval(() => setBlocked(v => v + (Math.random() > 0.8 ? 1 : 0)), 4500);
    return () => { clearInterval(t1); clearInterval(t2); };
  }, []);

  const servers = [
    { name: 'internal-api-server', tools: 8, team: 'Backend', status: 'فعال', risk: 'Low', invocations: 412, ok: true },
    { name: 'knowledge-base-mcp', tools: 5, team: 'Product', status: 'فعال', risk: 'Low', invocations: 288, ok: true },
    { name: 'data-pipeline-mcp', tools: 12, team: 'Data', status: 'هشدار', risk: 'Medium', invocations: 310, ok: null },
    { name: 'vendor-crm-mcp', tools: 4, team: 'Sales', status: 'فعال', risk: 'Low', invocations: 143, ok: true },
    { name: 'file-storage-mcp', tools: 6, team: 'Ops', status: 'خطا', risk: 'High', invocations: 87, ok: false },
  ];

  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl overflow-hidden">
      {/* Window bar */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-slate-200 dark:border-slate-800 bg-white/50 dark:bg-slate-900/50">
        <div className="w-3 h-3 rounded-full bg-rose-500/80" />
        <div className="w-3 h-3 rounded-full bg-amber-500/80" />
        <div className="w-3 h-3 rounded-full bg-emerald-500/80" />
        <div className="mx-auto flex items-center gap-2 px-3 py-1 rounded-md bg-slate-100 dark:bg-slate-800 text-xs text-slate-500 dark:text-slate-400 font-medium font-sans">
          <Server className="w-3 h-3" />
          Shahrah MCP Gateway
        </div>
      </div>

      <div className="p-5 space-y-4">
        {/* Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {[
            { label: 'MCP Server فعال', value: '۲۴', icon: Server, color: 'text-indigo-500', bg: 'bg-indigo-50 dark:bg-indigo-500/10' },
            { label: 'Tool Invocation امروز', value: invocations.toLocaleString('fa-IR'), icon: Wrench, color: 'text-emerald-500', bg: 'bg-emerald-50 dark:bg-emerald-500/10' },
            { label: 'درخواست Block شده', value: blocked.toLocaleString('fa-IR'), icon: XCircle, color: 'text-rose-500', bg: 'bg-rose-50 dark:bg-rose-500/10' },
            { label: 'Tool ثبت‌شده', value: '۱۸۰', icon: Package, color: 'text-purple-500', bg: 'bg-purple-50 dark:bg-purple-500/10' },
          ].map((s, i) => (
            <div key={i} className="p-3 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
              <div className="flex justify-between items-start mb-2">
                <span className="text-[10px] text-slate-500 dark:text-slate-400">{s.label}</span>
                <div className={`w-6 h-6 rounded-lg ${s.bg} flex items-center justify-center`}>
                  <s.icon className={`w-3 h-3 ${s.color}`} />
                </div>
              </div>
              <div className="text-lg font-bold font-sans text-slate-900 dark:text-white">{s.value}</div>
            </div>
          ))}
        </div>

        {/* MCP Server Registry + Auth Status */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {/* Server list */}
          <div className="md:col-span-2 p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <div className="flex justify-between items-center mb-3">
              <h3 className="text-xs font-bold text-slate-900 dark:text-white flex items-center gap-1.5">
                <Server className="w-3 h-3 text-indigo-500" /> MCP Server Registry
              </h3>
              <span className="flex items-center gap-1 text-[10px] text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-500/10 px-2 py-0.5 rounded-full font-medium">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse inline-block" />
                Live
              </span>
            </div>
            <div className="space-y-2 font-sans text-[10px]">
              {servers.map((srv, i) => (
                <div key={i} className="flex items-center gap-2 bg-slate-50 dark:bg-slate-800/40 rounded-lg px-3 py-2">
                  <Server className="w-3 h-3 text-slate-400 shrink-0" />
                  <span className="text-slate-700 dark:text-slate-300 font-medium flex-1 truncate font-sans">{srv.name}</span>
                  <span className="text-slate-400">{srv.tools} tools</span>
                  <span className="text-slate-500 hidden sm:block">{srv.team}</span>
                  <span className={`text-[9px] font-bold ${
                    srv.risk === 'Low' ? 'text-slate-400' :
                    srv.risk === 'Medium' ? 'text-amber-600 dark:text-amber-400' :
                    'text-rose-600 dark:text-rose-400'
                  }`}>{srv.risk}</span>
                  <span className="text-slate-400 font-sans">{srv.invocations}</span>
                  <span className={`w-5 h-5 rounded-full flex items-center justify-center shrink-0 ${
                    srv.ok === true ? 'bg-emerald-100 dark:bg-emerald-500/20' :
                    srv.ok === false ? 'bg-rose-100 dark:bg-rose-500/20' :
                    'bg-amber-100 dark:bg-amber-500/20'
                  }`}>
                    {srv.ok === true && <CheckCircle className="w-2.5 h-2.5 text-emerald-600 dark:text-emerald-400" />}
                    {srv.ok === false && <XCircle className="w-2.5 h-2.5 text-rose-600 dark:text-rose-400" />}
                    {srv.ok === null && <AlertTriangle className="w-2.5 h-2.5 text-amber-600 dark:text-amber-400" />}
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Auth + Policy Verdict */}
          <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 space-y-3">
            <h3 className="text-xs font-bold text-slate-900 dark:text-white flex items-center gap-1.5">
              <Key className="w-3 h-3 text-amber-500" /> Authentication
            </h3>
            <div className="space-y-1.5 font-sans text-[10px]">
              {[
                { method: 'OAuth 2.1', active: true },
                { method: 'API Token', active: true },
                { method: 'JWT Validation', active: true },
                { method: 'Header Auth', active: false },
                { method: 'Identity Forwarding', active: true },
              ].map((m, i) => (
                <div key={i} className="flex items-center justify-between py-1 border-b border-slate-50 dark:border-slate-800/50 last:border-0">
                  <span className="text-slate-600 dark:text-slate-400 font-sans">{m.method}</span>
                  <span className={`px-1.5 py-0.5 rounded-full font-bold text-[9px] ${
                    m.active
                      ? 'bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400'
                      : 'bg-slate-100 dark:bg-slate-700 text-slate-500'
                  }`}>{m.active ? 'فعال' : 'غیرفعال'}</span>
                </div>
              ))}
            </div>
            <div className="pt-1 border-t border-slate-100 dark:border-slate-800">
              <h3 className="text-[10px] font-bold text-slate-900 dark:text-white flex items-center gap-1.5 mb-2">
                <Shield className="w-3 h-3 text-purple-500" /> Policy Verdict
              </h3>
              {[
                { label: 'درخواست‌های Pass', val: '۱,۲۰۸', color: 'text-emerald-600 dark:text-emerald-400' },
                { label: 'Block شده', val: blocked.toLocaleString('fa-IR'), color: 'text-rose-600 dark:text-rose-400' },
                { label: 'Audit Logs', val: '۱,۲۴۰', color: 'text-slate-500' },
              ].map((r, i) => (
                <div key={i} className="flex justify-between items-center py-1 font-sans text-[10px]">
                  <span className="text-slate-500">{r.label}</span>
                  <span className={`font-bold ${r.color}`}>{r.val}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Audit Logs */}
        <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
          <h3 className="text-xs font-bold text-slate-900 dark:text-white mb-3 flex items-center gap-1.5">
            <BookOpen className="w-3 h-3 text-slate-500" /> Audit Logs — آخرین رویدادها
          </h3>
          <div className="overflow-x-auto">
            <table className="w-full text-[10px] font-sans text-right">
              <thead>
                <tr className="text-slate-400 dark:text-slate-500 border-b border-slate-100 dark:border-slate-800">
                  {['زمان', 'Agent', 'Tool', 'Server', 'وضعیت', 'Policy', 'Latency'].map(h => (
                    <th key={h} className="pb-2 font-medium">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {[
                  { time: '14:42:11', agent: 'support-bot', tool: 'search_kb', server: 'knowledge-base', status: 'Pass', policy: 'Allow', latency: '0.6s', ok: true },
                  { time: '14:42:05', agent: 'data-agent', tool: 'run_query', server: 'data-pipeline', status: 'Blocked', policy: 'Deny', latency: '0.1s', ok: false },
                  { time: '14:41:58', agent: 'product-bot', tool: 'get_context', server: 'internal-api', status: 'Pass', policy: 'Allow', latency: '0.4s', ok: true },
                  { time: '14:41:44', agent: 'ops-agent', tool: 'write_file', server: 'file-storage', status: 'Blocked', policy: 'RBAC', latency: '0.2s', ok: false },
                ].map((l, i) => (
                  <tr key={i} className="border-b border-slate-50 dark:border-slate-800/40 last:border-0">
                    <td className="py-2 text-slate-500">{l.time}</td>
                    <td className="py-2 text-slate-600 dark:text-slate-400">{l.agent}</td>
                    <td className="py-2 text-amber-600 dark:text-amber-400 font-medium">{l.tool}</td>
                    <td className="py-2 text-slate-500">{l.server}</td>
                    <td className="py-2">
                      <span className={`px-1.5 py-0.5 rounded-full font-bold text-[9px] ${
                        l.ok ? 'bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400'
                             : 'bg-rose-100 dark:bg-rose-500/20 text-rose-700 dark:text-rose-400'
                      }`}>{l.status}</span>
                    </td>
                    <td className="py-2 text-slate-500">{l.policy}</td>
                    <td className="py-2 text-slate-400">{l.latency}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}

// ─── Auth Methods Panel ───────────────────────────────────────────────────────
function AuthMethodsPanel() {
  const methods = [
    { name: 'OAuth 2.1', desc: 'جریان Authorization Code با PKCE', active: true, color: 'bg-indigo-100 dark:bg-indigo-500/20 text-indigo-700 dark:text-indigo-400' },
    { name: 'API Token', desc: 'توکن ثابت با سطح دسترسی مشخص', active: true, color: 'bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400' },
    { name: 'Header Auth', desc: 'ارسال Credential از طریق Header', active: true, color: 'bg-purple-100 dark:bg-purple-500/20 text-purple-700 dark:text-purple-400' },
    { name: 'JWT Validation', desc: 'اعتبارسنجی Token با Public Key', active: true, color: 'bg-amber-100 dark:bg-amber-500/20 text-amber-700 dark:text-amber-400' },
    { name: 'Identity Forwarding', desc: 'انتقال هویت کاربر به Server', active: true, color: 'bg-teal-100 dark:bg-teal-500/20 text-teal-700 dark:text-teal-400' },
    { name: 'Passthrough Headers', desc: 'عبور Header بدون تغییر', active: false, color: 'bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400' },
    { name: 'Custom Auth', desc: 'منطق احراز هویت اختصاصی', active: false, color: 'bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400' },
  ];
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs space-y-3">
      <div className="flex items-center gap-2 mb-1">
        <div className="w-7 h-7 rounded-lg bg-amber-100 dark:bg-amber-500/20 flex items-center justify-center">
          <Key className="w-3.5 h-3.5 text-amber-600 dark:text-amber-400" />
        </div>
        <span className="font-bold text-slate-800 dark:text-white">Authentication Methods</span>
      </div>
      <div className="space-y-2">
        {methods.map((m, i) => (
          <div key={i} className="flex items-center gap-3 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2.5">
            <div className={`px-2 py-0.5 rounded-full font-bold text-[9px] shrink-0 ${m.color}`}>{m.name}</div>
            <span className="flex-1 text-slate-500 dark:text-slate-400">{m.desc}</span>
            <span className={`w-2 h-2 rounded-full shrink-0 ${m.active ? 'bg-emerald-500 animate-pulse' : 'bg-slate-300 dark:bg-slate-600'}`} />
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Tool Registry Panel ──────────────────────────────────────────────────────
function ToolRegistryPanel() {
  const tools = [
    { server: 'knowledge-base-mcp', tools: ['search_docs', 'get_context', 'list_topics'], team: 'Product', risk: 'Low', access: 'همه تیم‌ها', invocations: 288 },
    { server: 'internal-api-server', tools: ['get_user', 'update_record', 'run_report'], team: 'Backend', risk: 'Medium', access: 'Backend, Ops', invocations: 412 },
    { server: 'data-pipeline-mcp', tools: ['run_query', 'export_csv', 'get_schema'], team: 'Data', risk: 'High', access: 'Data فقط', invocations: 310 },
  ];
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs space-y-3">
      <div className="flex items-center gap-2 mb-1">
        <div className="w-7 h-7 rounded-lg bg-purple-100 dark:bg-purple-500/20 flex items-center justify-center">
          <Package className="w-3.5 h-3.5 text-purple-600 dark:text-purple-400" />
        </div>
        <span className="font-bold text-slate-800 dark:text-white">Tool Registry</span>
        <span className="mr-auto text-[10px] text-purple-600 dark:text-purple-400 bg-purple-50 dark:bg-purple-500/10 px-2 py-0.5 rounded-full font-bold">۱۸۰ Tool</span>
      </div>
      <div className="space-y-3">
        {tools.map((srv, i) => (
          <div key={i} className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-3">
            <div className="flex items-center gap-2 mb-2">
              <Server className="w-3 h-3 text-indigo-400 shrink-0" />
              <span className="font-bold text-slate-700 dark:text-slate-300 font-sans flex-1 truncate">{srv.server}</span>
              <span className={`text-[9px] font-bold px-1.5 py-0.5 rounded-full ${
                srv.risk === 'Low' ? 'bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400' :
                srv.risk === 'Medium' ? 'bg-amber-100 dark:bg-amber-500/20 text-amber-700 dark:text-amber-400' :
                'bg-rose-100 dark:bg-rose-500/20 text-rose-700 dark:text-rose-400'
              }`}>{srv.risk}</span>
            </div>
            <div className="flex flex-wrap gap-1 mb-2">
              {srv.tools.map(t => (
                <span key={t} className="px-1.5 py-0.5 rounded bg-amber-50 dark:bg-amber-500/10 text-amber-700 dark:text-amber-400 font-mono font-bold text-[9px]">{t}</span>
              ))}
            </div>
            <div className="flex justify-between text-[9px] text-slate-400">
              <span>تیم: {srv.team}</span>
              <span>دسترسی: {srv.access}</span>
              <span className="font-sans">{srv.invocations} invocations</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Access Control Matrix ────────────────────────────────────────────────────
function AccessControlMatrix() {
  return (
    <div className="bg-slate-900 dark:bg-slate-950 border border-slate-800 rounded-2xl p-5 font-sans text-xs overflow-hidden">
      <div className="flex items-center gap-2 mb-4">
        <div className="w-7 h-7 rounded-lg bg-indigo-500/20 flex items-center justify-center">
          <Users className="w-3.5 h-3.5 text-indigo-400" />
        </div>
        <span className="font-bold text-white">Access Control Matrix — RBAC</span>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-right">
          <thead>
            <tr className="text-slate-500 border-b border-slate-800">
              {['تیم', 'Workspace', 'MCP Server', 'Toolهای مجاز', 'سطح دسترسی', 'Policy', 'وضعیت'].map(h => (
                <th key={h} className="pb-2 font-medium text-[9px] px-2">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {[
              { team: 'Product', ws: 'prod', server: 'knowledge-base', tools: 'همه', level: 'Read', policy: 'Allow', ok: true },
              { team: 'Backend', ws: 'api', server: 'internal-api', tools: 'انتخابی', level: 'Read/Write', policy: 'RBAC', ok: true },
              { team: 'Data', ws: 'analytics', server: 'data-pipeline', tools: 'محدود', level: 'Read', policy: 'Restricted', ok: null },
              { team: 'Sales', ws: 'crm', server: 'vendor-crm', tools: 'انتخابی', level: 'Read', policy: 'Allow', ok: true },
              { team: 'Ops', ws: 'infra', server: 'file-storage', tools: 'هیچ', level: 'None', policy: 'Deny', ok: false },
            ].map((r, i) => (
              <tr key={i} className="border-b border-slate-800/50 last:border-0 hover:bg-slate-800/30 transition-colors">
                <td className="py-2 px-2 font-bold text-white">{r.team}</td>
                <td className="py-2 px-2 text-slate-400 font-sans">{r.ws}</td>
                <td className="py-2 px-2 text-slate-400 font-sans truncate max-w-24">{r.server}</td>
                <td className="py-2 px-2 text-slate-300">{r.tools}</td>
                <td className="py-2 px-2 text-slate-300 font-sans">{r.level}</td>
                <td className="py-2 px-2">
                  <span className={`px-1.5 py-0.5 rounded-full text-[9px] font-bold font-sans ${
                    r.policy === 'Allow' ? 'bg-emerald-500/20 text-emerald-400' :
                    r.policy === 'Deny' ? 'bg-rose-500/20 text-rose-400' :
                    r.policy === 'RBAC' ? 'bg-indigo-500/20 text-indigo-400' :
                    'bg-amber-500/20 text-amber-400'
                  }`}>{r.policy}</span>
                </td>
                <td className="py-2 px-2">
                  <span className={`px-1.5 py-0.5 rounded-full text-[9px] font-bold ${
                    r.ok === true ? 'bg-emerald-500/20 text-emerald-400' :
                    r.ok === false ? 'bg-rose-500/20 text-rose-400' :
                    'bg-amber-500/20 text-amber-400'
                  }`}>{r.ok === true ? 'مجاز' : r.ok === false ? 'مسدود' : 'محدود'}</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ─── MCP Trace Viewer ─────────────────────────────────────────────────────────
function MCPTraceViewer() {
  const steps = [
    { step: 1, label: 'User Prompt', type: 'prompt', latency: '—', icon: Terminal, color: 'bg-slate-500', verdict: null },
    { step: 2, label: 'Agent Planning', type: 'planning', latency: '0.3s', icon: Bot, color: 'bg-indigo-500', verdict: null },
    { step: 3, label: 'MCP Tool Call', type: 'tool', latency: '0.6s', icon: Wrench, color: 'bg-amber-500', verdict: null },
    { step: 4, label: 'Tool Response', type: 'response', latency: '0.4s', icon: Database, color: 'bg-teal-500', verdict: null },
    { step: 5, label: 'Policy Check', type: 'policy', latency: '0.1s', icon: Shield, color: 'bg-purple-500', verdict: 'Pass' },
    { step: 6, label: 'Final Response', type: 'output', latency: '—', icon: CheckCircle, color: 'bg-emerald-500', verdict: null },
  ];
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs">
      <div className="flex items-center gap-2 mb-4">
        <div className="w-7 h-7 rounded-lg bg-indigo-100 dark:bg-indigo-500/20 flex items-center justify-center">
          <Activity className="w-3.5 h-3.5 text-indigo-600 dark:text-indigo-400" />
        </div>
        <span className="font-bold text-slate-800 dark:text-white">MCP Agent Trace — run-4412</span>
        <span className="mr-auto text-[10px] font-bold text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-500/10 px-2 py-0.5 rounded-full">Pass</span>
      </div>
      <div className="space-y-0">
        {steps.map((s, i) => (
          <div key={i} className="flex gap-3">
            <div className="flex flex-col items-center">
              <div className={`w-6 h-6 rounded-full ${s.color} flex items-center justify-center shrink-0 z-10`}>
                <s.icon className="w-3 h-3 text-white" />
              </div>
              {i < steps.length - 1 && <div className="w-px flex-1 bg-slate-200 dark:bg-slate-700 my-0.5" style={{ minHeight: '18px' }} />}
            </div>
            <div className="pb-3 flex-1">
              <div className="flex items-center gap-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg px-3 py-2">
                <span className="text-[9px] font-bold text-slate-400 w-4">{s.step}</span>
                <span className="font-medium text-slate-700 dark:text-slate-300 flex-1">{s.label}</span>
                <span className={`px-1.5 py-0.5 rounded text-[8px] font-bold font-sans ${
                  s.type === 'tool' ? 'bg-amber-100 dark:bg-amber-500/20 text-amber-700 dark:text-amber-400' :
                  s.type === 'policy' ? 'bg-purple-100 dark:bg-purple-500/20 text-purple-700 dark:text-purple-400' :
                  s.type === 'planning' ? 'bg-indigo-100 dark:bg-indigo-500/20 text-indigo-700 dark:text-indigo-400' :
                  'bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400'
                }`}>{s.type}</span>
                {s.latency !== '—' && <span className="text-slate-400 text-[9px]">{s.latency}</span>}
                {s.verdict && <span className="text-[9px] font-bold bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400 px-1.5 py-0.5 rounded-full">{s.verdict}</span>}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── MCP Analytics Dashboard ──────────────────────────────────────────────────
function MCPAnalyticsDashboard() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs space-y-3">
      <div className="flex items-center gap-2 mb-1">
        <div className="w-7 h-7 rounded-lg bg-emerald-100 dark:bg-emerald-500/20 flex items-center justify-center">
          <BarChart3 className="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
        </div>
        <span className="font-bold text-slate-800 dark:text-white">Analytics — مصرف MCP</span>
      </div>
      <div className="grid grid-cols-2 gap-2">
        {[
          { label: 'Request Volume', value: '۱,۲۴۰', bar: 80, color: 'bg-indigo-500', tc: 'text-indigo-600 dark:text-indigo-400' },
          { label: 'Error Rate', value: '۱.۳٪', bar: 13, color: 'bg-rose-500', tc: 'text-rose-600 dark:text-rose-400' },
          { label: 'Latency by Tool', value: '۰.۶s', bar: 35, color: 'bg-amber-500', tc: 'text-amber-600 dark:text-amber-400' },
          { label: 'Failed Invocations', value: '۱۶', bar: 8, color: 'bg-rose-500', tc: 'text-rose-600 dark:text-rose-400' },
        ].map((m, i) => (
          <div key={i} className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-2.5">
            <div className="flex justify-between mb-1.5">
              <span className="text-slate-500">{m.label}</span>
              <span className={`font-bold ${m.tc}`}>{m.value}</span>
            </div>
            <div className="h-1 bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden">
              <div className={`h-full ${m.color} rounded-full`} style={{ width: `${m.bar}%` }} />
            </div>
          </div>
        ))}
      </div>
      <div className="grid grid-cols-2 gap-2 text-[10px]">
        {[
          { label: 'Usage by Team', vals: [{ name: 'Backend', pct: '33٪' }, { name: 'Product', pct: '23٪' }, { name: 'Data', pct: '25٪' }, { name: 'Sales', pct: '19٪' }] },
          { label: 'Slow Tools', vals: [{ name: 'run_query', pct: '2.1s' }, { name: 'export_csv', pct: '1.8s' }, { name: 'write_file', pct: '1.4s' }] },
        ].map((g, i) => (
          <div key={i} className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-2.5">
            <div className="font-bold text-slate-600 dark:text-slate-400 mb-1.5">{g.label}</div>
            {g.vals.map((v, j) => (
              <div key={j} className="flex justify-between py-0.5 border-b border-slate-50 dark:border-slate-800/50 last:border-0">
                <span className="text-slate-500 font-sans">{v.name}</span>
                <span className="font-bold text-slate-700 dark:text-slate-300">{v.pct}</span>
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Runtime Policy Flow ──────────────────────────────────────────────────────
function RuntimePolicyFlow() {
  const steps = [
    { num: '۱', label: 'Agent درخواست Tool می‌دهد', icon: Bot, color: 'bg-indigo-500' },
    { num: '۲', label: 'بررسی هویت و دسترسی', icon: Key, color: 'bg-amber-500' },
    { num: '۳', label: 'اجرای Policy روی ورودی', icon: Shield, color: 'bg-purple-500' },
    { num: '۴', label: 'اجرای Tool (در صورت Pass)', icon: Wrench, color: 'bg-teal-500' },
  ];
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs space-y-3">
      <div className="flex items-center gap-2 mb-1">
        <div className="w-7 h-7 rounded-lg bg-purple-100 dark:bg-purple-500/20 flex items-center justify-center">
          <Settings className="w-3.5 h-3.5 text-purple-600 dark:text-purple-400" />
        </div>
        <span className="font-bold text-slate-800 dark:text-white">Runtime Policy Enforcement</span>
      </div>
      {steps.map((s, i) => (
        <div key={i}>
          <div className="flex items-center gap-3 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2.5">
            <div className={`w-6 h-6 rounded-full ${s.color} flex items-center justify-center shrink-0`}>
              <s.icon className="w-3 h-3 text-white" />
            </div>
            <span className="text-slate-700 dark:text-slate-300 font-medium">{s.label}</span>
            <span className="mr-auto text-[9px] text-slate-400">{s.num}</span>
          </div>
          {i < steps.length - 1 && (
            <div className="mr-5 border-r-2 border-dashed border-slate-200 dark:border-slate-700 h-3" />
          )}
        </div>
      ))}
      <div className="mt-2 p-3 rounded-xl bg-rose-50 dark:bg-rose-500/10 border border-rose-200 dark:border-rose-500/20">
        <div className="font-bold text-rose-700 dark:text-rose-400 mb-1.5 text-[10px]">در صورت Fail:</div>
        <div className="flex flex-wrap gap-1.5">
          {['Block', 'Redact', 'Route', 'Alert'].map(a => (
            <span key={a} className="px-2 py-0.5 rounded bg-rose-100 dark:bg-rose-500/20 text-rose-700 dark:text-rose-300 font-sans font-bold text-[9px]">{a}</span>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── Internal Tool Sharing Diagram ───────────────────────────────────────────
function InternalSharingDiagram() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs">
      <div className="flex items-center gap-2 mb-4">
        <div className="w-7 h-7 rounded-lg bg-teal-100 dark:bg-teal-500/20 flex items-center justify-center">
          <Lock className="w-3.5 h-3.5 text-teal-600 dark:text-teal-400" />
        </div>
        <span className="font-bold text-slate-800 dark:text-white">Credential Vault — بدون افشای Credential</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <div className="w-full px-4 py-2 rounded-xl bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-center text-slate-600 dark:text-slate-300 font-medium flex items-center justify-center gap-2">
          <Server className="w-3 h-3 text-indigo-500" /> Internal MCP Server
        </div>
        <ChevronRight className="w-4 h-4 text-slate-400 rotate-90" />
        <div className="px-6 py-3 rounded-xl bg-slate-900 border border-indigo-500/40 text-white font-bold flex items-center gap-2 shadow-[0_0_15px_rgba(99,102,241,0.15)] w-full justify-center">
          <Shield className="w-3.5 h-3.5 text-indigo-400" /> Shahrah MCP Gateway
          <span className="text-[9px] text-indigo-300 bg-indigo-500/20 px-1.5 py-0.5 rounded-full font-sans">Credential Vault</span>
        </div>
        <div className="grid grid-cols-3 gap-2 w-full text-[10px]">
          {[
            { label: 'Team A Agent', icon: Bot, color: 'bg-indigo-50 dark:bg-indigo-500/10 border-indigo-200 dark:border-indigo-500/30 text-indigo-700 dark:text-indigo-400' },
            { label: 'Team B Agent', icon: Bot, color: 'bg-purple-50 dark:bg-purple-500/10 border-purple-200 dark:border-purple-500/30 text-purple-700 dark:text-purple-400' },
            { label: 'Access Policy', icon: Lock, color: 'bg-emerald-50 dark:bg-emerald-500/10 border-emerald-200 dark:border-emerald-500/30 text-emerald-700 dark:text-emerald-400' },
          ].map((n, i) => (
            <div key={i} className={`p-2 rounded-lg border text-center flex flex-col items-center gap-1 ${n.color}`}>
              <n.icon className="w-3 h-3" />
              <span className="leading-tight">{n.label}</span>
            </div>
          ))}
        </div>
        <div className="w-full p-2.5 rounded-xl bg-emerald-50 dark:bg-emerald-500/10 border border-emerald-200 dark:border-emerald-500/20 text-center text-emerald-700 dark:text-emerald-400 font-bold text-[10px]">
          ✓ Credential هرگز در Agent یا کد پخش نمی‌شود
        </div>
      </div>
    </div>
  );
}

// ─── Org Scale Diagram ────────────────────────────────────────────────────────
function OrgScaleDiagram() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs">
      <div className="flex items-center gap-2 mb-4">
        <div className="w-7 h-7 rounded-lg bg-indigo-100 dark:bg-indigo-500/20 flex items-center justify-center">
          <Building className="w-3.5 h-3.5 text-indigo-600 dark:text-indigo-400" />
        </div>
        <span className="font-bold text-slate-800 dark:text-white">مقیاس سازمانی — یک Gateway</span>
      </div>
      <div className="space-y-2">
        {[
          { label: 'چند تیم', items: ['Backend', 'Product', 'Data', 'Sales', 'Ops'], color: 'text-indigo-500', bg: 'bg-indigo-50 dark:bg-indigo-500/10' },
          { label: 'چند MCP Server', items: ['knowledge-base', 'internal-api', 'data-pipeline', 'vendor-crm'], color: 'text-purple-500', bg: 'bg-purple-50 dark:bg-purple-500/10' },
        ].map((g, i) => (
          <div key={i} className="p-2.5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <div className={`text-[9px] font-bold ${g.color} mb-1.5`}>{g.label}</div>
            <div className="flex flex-wrap gap-1">
              {g.items.map(item => (
                <span key={item} className={`px-2 py-0.5 rounded-full text-[9px] font-medium ${g.bg} ${g.color}`}>{item}</span>
              ))}
            </div>
          </div>
        ))}
        <div className="p-2.5 rounded-xl bg-slate-900 border border-indigo-500/30 text-center">
          <div className="text-[10px] font-bold text-white flex items-center justify-center gap-1.5">
            <Shield className="w-3 h-3 text-indigo-400" /> Shahrah MCP Gateway
          </div>
          <div className="flex justify-center gap-2 mt-1.5">
            {['Identity Layer', 'Policy Layer'].map(l => (
              <span key={l} className="text-[9px] text-indigo-300 bg-indigo-500/20 px-2 py-0.5 rounded-full font-sans">{l}</span>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

// ─── CTA Mini Dashboard ───────────────────────────────────────────────────────
function CTAMiniDashboard() {
  const [invocations, setInvocations] = useState(1240);
  const [blocked, setBlocked] = useState(32);
  useEffect(() => {
    const t1 = setInterval(() => setInvocations(v => v + Math.floor(Math.random() * 2 + 1)), 1800);
    const t2 = setInterval(() => setBlocked(v => v + (Math.random() > 0.8 ? 1 : 0)), 4500);
    return () => { clearInterval(t1); clearInterval(t2); };
  }, []);
  return (
    <div className="bg-white/10 backdrop-blur-sm border border-white/20 rounded-2xl p-5 font-sans text-sm text-white">
      <div className="flex items-center gap-2 mb-4 pb-3 border-b border-white/10">
        <Server className="w-4 h-4 text-indigo-300" />
        <span className="text-xs font-bold text-white/80">MCP Gateway — وضعیت لحظه‌ای</span>
        <span className="mr-auto flex items-center gap-1 text-[10px] text-emerald-300 bg-emerald-500/20 px-2 py-0.5 rounded-full">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse inline-block" />
          Live
        </span>
      </div>
      <div className="grid grid-cols-2 gap-3">
        {[
          { label: '۲۴ MCP Server فعال', icon: Server, color: 'text-indigo-300' },
          { label: '۱۸۰ Tool ثبت‌شده', icon: Package, color: 'text-purple-300' },
          { label: `${invocations.toLocaleString('fa-IR')} Tool Invocation امروز`, icon: Wrench, color: 'text-emerald-300' },
          { label: '۹۸.۷٪ نرخ موفقیت', icon: CheckCircle, color: 'text-teal-300' },
          { label: `${blocked} درخواست Block شده`, icon: XCircle, color: 'text-rose-300' },
          { label: 'RBAC + Policy Enforcement', icon: Shield, color: 'text-amber-300' },
        ].map((m, i) => (
          <div key={i} className="flex items-center gap-2 bg-white/5 border border-white/10 rounded-xl p-2.5">
            <m.icon className={`w-3.5 h-3.5 shrink-0 ${m.color}`} />
            <span className="text-[11px] text-white/80 leading-tight">{m.label}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── FAQ ──────────────────────────────────────────────────────────────────────
const faqItems = [
  {
    question: 'MCP Gateway گُدار AI دقیقاً چه کاری انجام می‌دهد؟',
    answer: 'گُدار AI یک لایه مرکزی برای دسترسی Agentها به MCP Serverها و Toolها فراهم می‌کند. این لایه Authentication، Access Control، Tool Registry، Observability، Audit Logs، Analytics و Policy Enforcement را مدیریت می‌کند.',
  },
  {
    question: 'آیا برای استفاده از MCP Gateway باید Agentها یا MCP Serverها را تغییر دهیم؟',
    answer: 'هدف MCP Gateway این است که بدون تغییر جدی در Agentها یا Serverها، یک لایه کنترل، امنیت و مشاهده‌پذیری روی مسیر MCP اضافه کند.',
  },
  {
    question: 'آیا می‌توانیم دسترسی به Toolها را در سطح تیم یا کاربر کنترل کنیم؟',
    answer: 'بله. با RBAC و Access Control می‌توانید مشخص کنید هر سازمان، Workspace، تیم، کاربر یا Agent به کدام MCP Server و کدام Tool دسترسی دارد.',
  },
  {
    question: 'اگر Agent بخواهد Tool غیرمجاز را اجرا کند چه می‌شود؟',
    answer: 'اگر Tool یا Server در Gateway تأیید نشده باشد یا Policy اجازه ندهد، درخواست می‌تواند قبل از اجرا Block شود و در Audit Logs ثبت شود.',
  },
  {
    question: 'آیا اجرای Toolها قابل Trace و Debug است؟',
    answer: 'بله. می‌توانید ببینید هر Agent چه Toolی را چه زمانی، با چه ورودی، چه خروجی، چه Latency، چه خطا و چه Policy Verdictی اجرا کرده است.',
  },
  {
    question: 'آیا می‌توانیم Toolهای داخلی و Vendor Toolها را کنار هم مدیریت کنیم؟',
    answer: 'بله. گُدار AI می‌تواند Toolهای داخلی، Vendor Toolها و Partner Toolها را در یک Registry و تحت یک مدل Governance، Access Control، Observability و Policy مدیریت کند.',
  },
  {
    question: 'آیا برای استفاده از گُدار AI نیاز به پرداخت ارزی دارم؟',
    answer: 'خیر. گُدار AI برای تیم‌های ایرانی طراحی شده و هدف آن ساده‌کردن دسترسی به مدل‌ها، Agentها و Toolهای AI بدون نیاز به پرداخت ارزی مستقیم و بدون درگیری با محدودیت‌های رایج سرویس‌های خارجی است.',
  },
];

// ─── Main Page ────────────────────────────────────────────────────────────────
export function MCPGateway() {
  return (
    <main className="flex min-h-screen flex-col bg-white dark:bg-[#0B0F19] text-slate-900 dark:text-white" dir="rtl">

      {/* ═══ 1. Hero ══════════════════════════════════════════════════════════ */}
      <section className="relative pt-32 pb-20 overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,#4F46E5_0%,transparent_50%)] opacity-[0.08] dark:opacity-[0.15] pointer-events-none" />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="text-center max-w-4xl mx-auto">
            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5 }}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-indigo-50 dark:bg-indigo-500/10 border border-indigo-100 dark:border-indigo-500/20 text-indigo-600 dark:text-indigo-400 text-sm font-bold mb-8"
              style={{ fontFamily: 'var(--font-heading)' }}
            >
              <Server className="w-4 h-4" />
              MCP Gateway برای تیم‌های ایرانی AI
            </motion.div>

            <motion.h1
              className="text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight mb-8 leading-tight text-slate-900 dark:text-white"
              style={{ fontFamily: 'var(--font-heading)' }}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
            >
              MCP را{' '}
              <span className="text-transparent bg-clip-text bg-gradient-to-l from-indigo-500 to-purple-500">
                امن، قابل‌مشاهده
              </span>{' '}
              و آماده Production اجرا کنید
            </motion.h1>

            <motion.p
              className="text-lg md:text-xl text-slate-600 dark:text-slate-400 mb-10 leading-relaxed max-w-3xl mx-auto"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
            >
              با گُدار AI، Agentها را به MCP Serverها و Toolهای داخلی یا خارجی متصل کنید، بدون اینکه دسترسی‌ها پراکنده شوند یا Credentialها در تیم‌ها پخش شوند. از یک لایه مرکزی، Authentication، Access Control، Observability، Policy Enforcement و مدیریت مصرف MCP را کنترل کنید؛ بدون درگیری با تحریم یا پرداخت ارزی مستقیم.
            </motion.p>

            <motion.div
              className="flex flex-col sm:flex-row items-center justify-center gap-4"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
            >
              <Link href="/signup" className="w-full sm:w-auto px-8 py-4 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-lg shadow-[0_0_20px_rgba(79,70,229,0.3)] hover:shadow-[0_0_30px_rgba(79,70,229,0.5)] transition-all active:scale-95 flex items-center justify-center gap-2 group">
                شروع کنید
                <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
              </Link>
              <Link href="/contact" className="w-full sm:w-auto px-8 py-4 bg-white dark:bg-slate-900 text-slate-900 dark:text-white border-2 border-slate-200 dark:border-slate-800 hover:border-indigo-600 dark:hover:border-indigo-500 rounded-xl font-bold text-lg transition-all active:scale-95 flex items-center justify-center">
                درخواست دمو
              </Link>
            </motion.div>
          </div>

          <motion.div
            className="mt-20 relative mx-auto w-full max-w-6xl"
            initial={{ opacity: 0, y: 40 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, delay: 0.3 }}
          >
            <div className="absolute inset-0 bg-gradient-to-t from-white via-transparent to-transparent dark:from-[#0B0F19] z-20 pointer-events-none" style={{ top: '65%' }} />
            <MCPGatewayDashboard />
          </motion.div>
        </div>
      </section>

      {/* ═══ 2. Trust Bar ════════════════════════════════════════════════════ */}
      <section className="py-12 border-y border-slate-100 dark:border-slate-800/50 bg-slate-50/50 dark:bg-slate-900/20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-8">
            <h2 className="text-xl font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
              ساخته‌شده برای تیم‌هایی که Agentها را به Toolهای واقعی متصل می‌کنند
            </h2>
            <p className="text-sm text-slate-600 dark:text-slate-400 max-w-2xl mx-auto">
              گُدار AI به تیم‌های Backend، Platform، AI و امنیت کمک می‌کند MCP Serverها و Toolها را در یک لایه مرکزی مدیریت کنند؛ با دسترسی کنترل‌شده، رهگیری کامل، Policyهای اجرایی، مانیتورینگ مصرف و تجربه‌ای مناسب بازار ایران.
            </p>
          </div>
          <div className="flex flex-wrap justify-center gap-3">
            {[
              'استارتاپ‌های AI', 'تیم‌های Backend و Platform', 'تیم‌های امنیت و DevOps',
              'محصولات SaaS', 'تیم‌های R&D و AI Lab', 'سازمان‌های داده‌محور',
              'تیم‌هایی که Agentهای متصل به Tool می‌سازند',
              'شرکت‌هایی که ابزارهای داخلی را به AI متصل می‌کنند',
            ].map((chip, idx) => (
              <div key={idx} className="px-4 py-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-full text-sm font-medium text-slate-700 dark:text-slate-300 shadow-sm">
                {chip}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 3. MCP Control Plane ════════════════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              لایه کنترل مرکزی برای MCP
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              MCP وقتی در محصول واقعی استفاده می‌شود، فقط یک اتصال ساده بین Agent و Tool نیست. باید بدانید چه Agentی به چه Toolی دسترسی دارد، چه کسی احراز هویت شده، کدام Tool اجرا شده و کدام Policy روی مسیر اعمال شده است.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Card 1 — Auth */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(245,158,11,0.07)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-amber-100 dark:bg-amber-500/20 text-amber-600 dark:text-amber-400 rounded-xl flex items-center justify-center mb-6">
                  <Key className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  پشتیبانی داخلی از انواع Authentication
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  گُدار AI می‌تواند مسیر اتصال به MCP Serverها را با روش‌های مختلف Authentication مدیریت کند؛ از OAuth 2.1 و API Token تا Header Auth، JWT Validation، Identity Forwarding و Passthrough Headers. تیم شما می‌تواند ساختار Auth مناسب خود را داشته باشد، بدون اینکه Agentها یا Serverها را پیچیده‌تر کند.
                </p>
                <p className="text-xs text-amber-600 dark:text-amber-400 font-medium mb-6">
                  اتصال امن به Toolها، بدون پخش شدن Credentialها در کد و تیم‌ها.
                </p>
                <AuthMethodsPanel />
              </div>
            </MagicCard>

            {/* Card 2 — Registry */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(139,92,246,0.07)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-purple-100 dark:bg-purple-500/20 text-purple-600 dark:text-purple-400 rounded-xl flex items-center justify-center mb-6">
                  <Package className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  Registry مرکزی برای MCP Serverها و Toolها
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  تمام MCP Serverها و Toolهای قابل‌استفاده را در یک Registry واحد ببینید. تیم‌ها می‌توانند Toolهای تأییدشده را پیدا کنند، Serverهای جدید اضافه کنند و دسترسی‌ها را بدون ایجاد پراکندگی در سازمان مدیریت کنند.
                </p>
                <p className="text-xs text-purple-600 dark:text-purple-400 font-medium mb-6">
                  اگر Tool در Gateway تأیید نشده باشد، Agent اجازه استفاده از آن را ندارد.
                </p>
                <ToolRegistryPanel />
              </div>
            </MagicCard>

            {/* Card 3 — Access Control */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(99,102,241,0.07)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-indigo-100 dark:bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 rounded-xl flex items-center justify-center mb-6">
                  <Users className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  کنترل دسترسی دقیق برای Serverها و Toolها
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  با RBAC، دسترسی را در سطح سازمان، Workspace، تیم یا کاربر مشخص کنید. هر تیم فقط به Toolهایی دسترسی داشته باشد که واقعاً نیاز دارد؛ بدون Shadow Tooling و بدون اشتراک‌گذاری مستقیم Credentialها.
                </p>
                <p className="text-xs text-indigo-600 dark:text-indigo-400 font-medium mb-6">
                  Toolها را بین تیم‌ها به اشتراک بگذارید، نه Credentialها را.
                </p>
                <AccessControlMatrix />
              </div>
            </MagicCard>

            {/* Card 4 — Observability */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(99,102,241,0.07)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-indigo-100 dark:bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 rounded-xl flex items-center justify-center mb-6">
                  <Eye className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  Observability کامل برای Agent Workflowها
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  برای هر Agent ببینید کدام MCP Tool را صدا زده، با چه ورودی، چه خروجی، چه Latency، چه خطا و چه نتیجه Policy. هر Tool Invocation لاگ، Audit و قابل جست‌وجو باشد تا Debug و بررسی ریسک سریع‌تر انجام شود.
                </p>
                <p className="text-xs text-indigo-600 dark:text-indigo-400 font-medium mb-6">
                  Agent فقط پاسخ نمی‌دهد؛ Tool اجرا می‌کند. اجرای Toolها باید قابل‌ردیابی باشد.
                </p>
                <MCPTraceViewer />
              </div>
            </MagicCard>
          </div>

          {/* Cards 5 & 6 */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mt-8">
            {/* Card 5 — Analytics */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(16,185,129,0.06)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-emerald-100 dark:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 rounded-xl flex items-center justify-center mb-6">
                  <BarChart3 className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  Analytics مصرف و Performance برای MCP
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  حجم درخواست، Latency، خطاها، Toolهای کند، Serverهای ناپایدار و الگوهای مصرف را در سطح Tool، Server، تیم، کاربر یا Workspace تحلیل کنید. قبل از اینکه مشکل یک Server روی Agentها اثر بگذارد، آن را شناسایی کنید.
                </p>
                <p className="text-xs text-emerald-600 dark:text-emerald-400 font-medium mb-6">
                  بدانید کدام Tool واقعاً استفاده می‌شود، کدام کند است و کجا خطا زیاد شده است.
                </p>
                <MCPAnalyticsDashboard />
              </div>
            </MagicCard>

            {/* Card 6 — Policy Enforcement */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(139,92,246,0.07)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-purple-100 dark:bg-purple-500/20 text-purple-600 dark:text-purple-400 rounded-xl flex items-center justify-center mb-6">
                  <Settings className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  Policy Enforcement در زمان اجرا
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  Policyها فقط برای گزارش نیستند. گُدار AI می‌تواند قبل از اجرای Tool، دسترسی غیرمجاز را متوقف کند، PII را Redact کند، محتوای حساس را فیلتر کند یا درخواست‌های مشکوک را Block کند. این Policyها می‌توانند در مسیر LLM و MCP به‌صورت یکپارچه اجرا شوند.
                </p>
                <p className="text-xs text-purple-600 dark:text-purple-400 font-medium mb-6">
                  قبل از اجرای Tool، تصمیم بگیرید اجازه دارد یا نه.
                </p>
                <RuntimePolicyFlow />
              </div>
            </MagicCard>
          </div>
        </div>
      </section>

      {/* ═══ 4. One Interface ════════════════════════════════════════════════ */}
      <section className="py-24 bg-slate-900 dark:bg-[#060913] text-white border-y border-slate-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              یک Interface واحد برای تمام MCP Serverهای شما
            </h2>
            <p className="text-lg text-slate-400">
              MCP Serverها می‌توانند داخلی، Vendor-based یا Partner-based باشند. گُدار AI کمک می‌کند همه آن‌ها از یک مسیر امن، قابل‌کنترل و قابل‌مشاهده در اختیار Agentها قرار بگیرند.
            </p>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Capability 1 — Internal Sharing */}
            <div className="bg-slate-800/50 border border-slate-700 rounded-2xl p-6 flex flex-col">
              <div className="w-10 h-10 rounded-xl bg-teal-500/20 text-teal-400 flex items-center justify-center mb-4">
                <Lock className="w-5 h-5" />
              </div>
              <h3 className="font-bold text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                اشتراک‌گذاری Toolهای داخلی بدون افشای Credentialها
              </h3>
              <p className="text-slate-400 text-sm leading-relaxed mb-5 flex-1">
                Toolها و MCP Serverهای داخلی را برای Agentها و تیم‌ها قابل‌استفاده کنید، اما Credentialها را در Gateway نگه دارید. کاربران و Agentها از مسیر کنترل‌شده احراز هویت می‌شوند و مجوزها در سطح تیم یا کاربر اعمال می‌شوند.
              </p>
              <InternalSharingDiagram />
            </div>

            {/* Capability 2 — Scale */}
            <div className="bg-slate-800/50 border border-slate-700 rounded-2xl p-6 flex flex-col">
              <div className="w-10 h-10 rounded-xl bg-indigo-500/20 text-indigo-400 flex items-center justify-center mb-4">
                <Building className="w-5 h-5" />
              </div>
              <h3 className="font-bold text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                مقیاس‌دادن دسترسی MCP در کل سازمان
              </h3>
              <p className="text-slate-400 text-sm leading-relaxed mb-5 flex-1">
                به‌جای اینکه هر تیم مسیر اتصال جداگانه بسازد، همه MCP Serverها را از طریق گُدار AI در اختیار سازمان قرار دهید. کاربران با هویت سازمانی احراز هویت شوند، و تمام ترافیک از Gateway عبور کند.
              </p>
              <OrgScaleDiagram />
            </div>

            {/* Capability 3 — Vendor */}
            <div className="bg-slate-800/50 border border-slate-700 rounded-2xl p-6 flex flex-col">
              <div className="w-10 h-10 rounded-xl bg-purple-500/20 text-purple-400 flex items-center justify-center mb-4">
                <ShieldCheck className="w-5 h-5" />
              </div>
              <h3 className="font-bold text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                Vendor Toolها زیر Governance شما
              </h3>
              <p className="text-slate-400 text-sm leading-relaxed mb-5 flex-1">
                Toolهای Vendor یا Partner را با OAuth، API Key یا Custom Auth متصل کنید و آن‌ها را تحت همان مدل Governance، Access Control، Observability و Policy Enforcement ابزارهای داخلی قرار دهید.
              </p>
              <div className="grid grid-cols-2 gap-2 font-sans text-xs">
                {['Vendor MCP Server', 'Partner Tool', 'External Data Tool', 'SaaS Connector', 'Internal API Tool', 'Knowledge Tool'].map(t => (
                  <div key={t} className="flex items-center gap-2 p-2 rounded-lg bg-slate-900/60 border border-slate-700">
                    <div className="w-5 h-5 rounded bg-slate-700 flex items-center justify-center shrink-0">
                      <Wrench className="w-3 h-3 text-slate-400" />
                    </div>
                    <span className="text-slate-300 text-[10px] leading-tight">{t}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ═══ 5. What You Can Build ════════════════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              با MCP Gateway گُدار AI چه چیزی می‌توانید بسازید؟
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              MCP Gateway به شما اجازه می‌دهد Toolها را از چند Server مختلف به Agentها متصل کنید، اما دسترسی، امنیت، مصرف و Policyها را در یک لایه مرکزی نگه دارید.
            </p>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            {[
              {
                icon: Bot, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10',
                title: 'Agentهای متصل به ابزارهای داخلی',
                desc: 'Agentهایی بسازید که به سیستم‌های داخلی، APIها، دیتابیس‌ها، فایل‌ها یا Knowledge Base سازمان متصل می‌شوند، اما هر Tool Call آن‌ها قابل‌ردیابی و کنترل‌شده است.',
              },
              {
                icon: ShieldCheck, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10',
                title: 'دستیارهای سازمانی امن',
                desc: 'برای تیم‌های Support، Sales، Finance، Product یا Ops دستیارهایی بسازید که فقط به Toolهای مجاز همان تیم دسترسی دارند.',
              },
              {
                icon: LayoutGrid, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10',
                title: 'Tool Marketplace داخلی',
                desc: 'یک Registry داخلی از Toolهای تأییدشده ایجاد کنید تا تیم‌ها بتوانند ابزارهای قابل‌اعتماد را پیدا، تست و در Agentهای خود استفاده کنند.',
              },
              {
                icon: GitBranch, color: 'text-teal-600 dark:text-teal-400', bg: 'bg-teal-50 dark:bg-teal-500/10',
                title: 'Multi-agent Workflowهای قابل‌ممیزی',
                desc: 'وقتی چند Agent به چند Tool متصل هستند، اجرای هر مرحله، هر Tool Invocation و هر Policy Decision را Trace و Audit کنید.',
              },
              {
                icon: Network, color: 'text-rose-600 dark:text-rose-400', bg: 'bg-rose-50 dark:bg-rose-500/10',
                title: 'اتصال کنترل‌شده به Vendorها',
                desc: 'Agentها را به Toolهای بیرونی وصل کنید، اما Credential، سطح دسترسی، Rate Limit و Policyها را در اختیار Gateway نگه دارید.',
              },
              {
                icon: Database, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10',
                title: 'AI Context Platform برای سازمان',
                desc: 'زیرساختی بسازید که Agentها از طریق آن به Context، ابزارها، داده‌ها و MCP Serverهای سازمان دسترسی امن و قابل‌ردیابی داشته باشند.',
              },
            ].map((card, i) => (
              <motion.div
                key={i}
                className="bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 hover:border-indigo-200 dark:hover:border-indigo-800 transition-colors group"
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: (i % 3) * 0.1 }}
              >
                <div className={`w-10 h-10 ${card.bg} ${card.color} rounded-xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform`}>
                  <card.icon className="w-5 h-5" />
                </div>
                <h3 className="font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                  {card.title}
                </h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">{card.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 6. Enterprise Platform ══════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              پلتفرم Context سازمانی برای AI بسازید
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              وقتی Agentها به داده‌ها و ابزارهای واقعی وصل می‌شوند، MCP به یک لایه کلیدی در AI stack تبدیل می‌شود. گُدار AI کمک می‌کند این لایه را امن، قابل‌ممیزی و قابل‌مقیاس اجرا کنید.
            </p>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {[
              {
                icon: ShieldCheck, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10', border: 'border-indigo-200 dark:border-indigo-500/20',
                title: 'Secure by Design',
                desc: 'دسترسی به Toolها، Serverها و Credentialها از ابتدا با کنترل امنیتی طراحی می‌شود؛ با مدیریت هویت، سطح دسترسی، Audit Logs و Policyهای Runtime.',
                tags: ['RBAC', 'Audit Logs', 'Runtime Policy', 'Credential Vault'],
              },
              {
                icon: Layers, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10', border: 'border-purple-200 dark:border-purple-500/20',
                title: 'یک Control Plane برای LLM و MCP',
                desc: 'مدل‌ها و Toolها را جدا از هم مدیریت نکنید. گُدار AI می‌تواند Gateway، Observability و Policy را برای هر دو مسیر LLM و MCP در یک تجربه واحد فراهم کند.',
                tags: ['LLM Gateway', 'MCP Gateway', 'Unified Observability'],
              },
              {
                icon: Settings, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10', border: 'border-emerald-200 dark:border-emerald-500/20',
                title: 'قابل‌استقرار متناسب با نیاز تیم شما',
                desc: 'بسته به نیاز تیم، معماری گُدار AI می‌تواند برای SaaS، Private Cloud، VPC یا Self-hosted در نظر گرفته شود تا کنترل و امنیت بیشتری فراهم شود.',
                tags: ['SaaS', 'Private Cloud', 'VPC', 'Self-hosted'],
              },
            ].map((card, i) => (
              <MagicCard
                key={i}
                className={`bg-white dark:bg-slate-900/60 border ${card.border} dark:border-slate-800 rounded-2xl overflow-hidden`}
                gradientColor="rgba(99,102,241,0.07)"
              >
                <div className="p-6">
                  <div className={`w-12 h-12 ${card.bg} ${card.color} rounded-xl flex items-center justify-center mb-5`}>
                    <card.icon className="w-6 h-6" />
                  </div>
                  <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                    {card.title}
                  </h3>
                  <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed mb-4">{card.desc}</p>
                  <div className="flex flex-wrap gap-1.5">
                    {card.tags.map(tag => (
                      <span key={tag} className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${card.bg} ${card.color} font-sans`}>{tag}</span>
                    ))}
                  </div>
                </div>
              </MagicCard>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 7. Differentiation ══════════════════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,#4F46E5_0%,transparent_60%)] opacity-[0.04] dark:opacity-[0.08] pointer-events-none" />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              چرا MCP Gateway گُدار AI برای تیم‌های ایرانی مهم‌تر است؟
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              برای تیم‌های ایرانی، اتصال Agentها به Toolهای واقعی فقط مسئله فنی نیست؛ دسترسی پایدار، تحریم، پرداخت، امنیت داده، Credential Management و کنترل ابزارهای داخلی هم اهمیت دارد.
            </p>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {[
              {
                icon: Shield, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10', border: 'border-indigo-100 dark:border-indigo-500/20',
                title: 'بدون محدودیت‌های تحریمی',
                desc: 'با گُدار AI، تیم شما برای اتصال Agentها به مدل‌ها و ابزارهای AI خارجی کمتر درگیر محدودیت‌های تحریم، مسدود شدن حساب یا قطع دسترسی مستقیم می‌شود.',
              },
              {
                icon: CreditCard, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10', border: 'border-emerald-100 dark:border-emerald-500/20',
                title: 'بدون پرداخت ارزی مستقیم',
                desc: 'به‌جای درگیری با کارت اعتباری خارجی، پرداخت دلاری و حساب‌های بین‌المللی، مصرف مدل‌ها، Agentها و ابزارهای MCP را از طریق گُدار AI و متناسب با نیاز بازار ایران مدیریت کنید.',
              },
              {
                icon: Lock, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10', border: 'border-purple-100 dark:border-purple-500/20',
                title: 'کنترل متمرکز روی Toolها، Credentialها و Policyها',
                desc: 'به‌جای اتصال پراکنده Agentها به چند Server و Tool، گُدار AI یک مسیر مرکزی برای MCP Registry، Access Control، Observability، Audit و Policy Enforcement فراهم می‌کند.',
              },
              {
                icon: Globe, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10', border: 'border-amber-100 dark:border-amber-500/20',
                title: 'مناسب برای داده فارسی و ابزارهای داخلی ایرانی',
                desc: 'گُدار AI می‌تواند برای سناریوهایی مثل داده فارسی، سیستم‌های داخلی، شماره تماس، کد ملی، منابع سازمانی، APIهای داخلی و Workflowهای تیم‌های ایرانی تنظیم شود.',
              },
            ].map((diff, i) => (
              <MagicCard
                key={i}
                className={`bg-white dark:bg-slate-900/50 border ${diff.border} dark:border-slate-800 rounded-3xl p-8 flex gap-6 items-start`}
                gradientColor="rgba(99,102,241,0.08)"
              >
                <div className={`w-14 h-14 ${diff.bg} ${diff.color} rounded-2xl flex items-center justify-center shrink-0`}>
                  <diff.icon className="w-7 h-7" />
                </div>
                <div>
                  <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                    {diff.title}
                  </h3>
                  <p className="text-slate-600 dark:text-slate-400 leading-relaxed">{diff.desc}</p>
                </div>
              </MagicCard>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 8. Use Cases ════════════════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              MCP Gateway گُدار AI برای چه سناریوهایی مناسب است؟
            </h2>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            {[
              {
                icon: Bot, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10',
                title: 'Agentهای متصل به ابزارهای سازمانی',
                desc: 'Agentهایی بسازید که به ابزارهای داخلی، APIها، فایل‌ها، Knowledge Base یا سیستم‌های سازمانی متصل می‌شوند، اما دسترسی و اجرای آن‌ها قابل‌کنترل است.',
              },
              {
                icon: Layers, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10',
                title: 'تیم‌های Backend و Platform',
                desc: 'MCP Serverها، Tool Registry، Credentialها، Auth، Policyها و Observability را از یک لایه مرکزی مدیریت کنید.',
              },
              {
                icon: ShieldCheck, color: 'text-rose-600 dark:text-rose-400', bg: 'bg-rose-50 dark:bg-rose-500/10',
                title: 'تیم‌های امنیت و DevOps',
                desc: 'دسترسی غیرمجاز به Toolها را متوقف کنید، Audit Logs داشته باشید، Policyهای Runtime اجرا کنید و Credentialها را از کد و تیم‌ها جدا نگه دارید.',
              },
              {
                icon: Activity, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10',
                title: 'تیم‌های محصول و AI',
                desc: 'Agentهای محصولی را به Toolهای مختلف متصل کنید، اما کیفیت، خطا، Latency، استفاده از Tool و Policy Verdictها را مشاهده کنید.',
              },
              {
                icon: Building, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10',
                title: 'سازمان‌های دارای چند تیم AI',
                desc: 'اگر چند تیم در سازمان از Agent و Tool استفاده می‌کنند، با Registry و RBAC از Shadow Tooling، دسترسی‌های پراکنده و Credential Sharing جلوگیری کنید.',
              },
              {
                icon: Network, color: 'text-teal-600 dark:text-teal-400', bg: 'bg-teal-50 dark:bg-teal-500/10',
                title: 'اتصال کنترل‌شده به Vendor Toolها',
                desc: 'Toolهای بیرونی را به Agentها متصل کنید، اما Auth، Rate Limit، دسترسی، Policy و لاگ‌ها را در Gateway مرکزی مدیریت کنید.',
              },
            ].map((uc, i) => (
              <motion.div
                key={i}
                className="bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 hover:border-indigo-200 dark:hover:border-indigo-800 transition-colors group"
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: (i % 3) * 0.1 }}
              >
                <div className={`w-10 h-10 ${uc.bg} ${uc.color} rounded-xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform`}>
                  <uc.icon className="w-5 h-5" />
                </div>
                <h3 className="font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                  {uc.title}
                </h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">{uc.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 9. Resources ════════════════════════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              راهنماها و منابع پیشنهادی
            </h2>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {[
              {
                tag: 'راهنما', color: 'text-indigo-600 dark:text-indigo-400', tagBg: 'bg-indigo-50 dark:bg-indigo-500/10',
                title: 'چرا MCP Gateway برای Agentهای Production ضروری است؟',
                desc: 'وقتی Agentها به Toolهای واقعی متصل می‌شوند، Auth، Access Control، Observability و Policy Enforcement به بخشی حیاتی از زیرساخت تبدیل می‌شوند.',
              },
              {
                tag: 'مقاله', color: 'text-purple-600 dark:text-purple-400', tagBg: 'bg-purple-50 dark:bg-purple-500/10',
                title: 'چالش پنهان MCP در سازمان‌ها',
                desc: 'با زیاد شدن MCP Serverها، ابزارها و تیم‌ها، Sprawl، Shadow Tooling و Credential Sharing می‌تواند امنیت و مدیریت را سخت کند.',
              },
              {
                tag: 'راهنما', color: 'text-emerald-600 dark:text-emerald-400', tagBg: 'bg-emerald-50 dark:bg-emerald-500/10',
                title: 'چطور Tool Registry داخلی برای AI بسازیم؟',
                desc: 'با Registry مرکزی، تیم‌ها می‌توانند Toolهای تأییدشده را پیدا کنند، دسترسی بگیرند و بدون ساخت اتصال‌های پراکنده از آن‌ها استفاده کنند.',
              },
            ].map((res, i) => (
              <div key={i} className="bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 hover:shadow-lg hover:border-slate-300 dark:hover:border-slate-700 transition-all cursor-pointer group flex flex-col">
                <span className={`text-xs font-bold px-2.5 py-1 rounded-full w-fit mb-4 ${res.tagBg} ${res.color}`}>{res.tag}</span>
                <h3 className="font-bold text-slate-900 dark:text-white mb-3 leading-snug flex-1" style={{ fontFamily: 'var(--font-heading)' }}>
                  {res.title}
                </h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed mb-4">{res.desc}</p>
                <div className={`flex items-center gap-1 text-sm font-bold ${res.color} group-hover:gap-2 transition-all`}>
                  بیشتر بخوانید
                  <ArrowLeft className="w-4 h-4" />
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 10. FAQ ══════════════════════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-t border-slate-100 dark:border-slate-800/50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white" style={{ fontFamily: 'var(--font-heading)' }}>
              سوالات متداول
            </h2>
          </div>
          <FAQAccordion items={faqItems} />
        </div>
      </section>

      {/* ═══ 11. Final CTA ════════════════════════════════════════════════════ */}
      <section className="py-24 relative overflow-hidden">
        <div className="absolute inset-0 bg-indigo-600 dark:bg-indigo-900/40" />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            <div>
              <h2 className="text-3xl md:text-4xl font-bold text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
                MCP-powered Agentهای خود را با اطمینان به Production ببرید
              </h2>
              <p className="text-lg text-indigo-200 mb-10 leading-relaxed">
                با گُدار AI، MCP Serverها، Toolها، Credentialها، Access Control، Observability، Audit Logs و Policyهای Runtime را از یک لایه مرکزی مدیریت کنید؛ بدون محدودیت‌های تحریمی و بدون نیاز به پرداخت ارزی مستقیم.
              </p>
              <div className="flex flex-col sm:flex-row items-start gap-4">
                <Link href="/signup" className="w-full sm:w-auto px-8 py-4 bg-white text-indigo-700 hover:bg-indigo-50 rounded-xl font-bold text-lg shadow-[0_0_30px_rgba(0,0,0,0.2)] transition-all active:scale-95 flex items-center justify-center gap-2 group">
                  شروع کنید
                  <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
                </Link>
                <Link href="/contact" className="w-full sm:w-auto px-8 py-4 bg-white/10 hover:bg-white/20 text-white border border-white/20 hover:border-white/40 rounded-xl font-bold text-lg transition-all active:scale-95 flex items-center justify-center">
                  درخواست دمو
                </Link>
              </div>
              <div className="mt-8 flex flex-wrap gap-3">
                {[
                  { icon: Shield, label: 'بدون تحریم' },
                  { icon: CreditCard, label: 'بدون پرداخت ارزی' },
                  { icon: ShieldCheck, label: 'آماده Production' },
                ].map((b, i) => (
                  <div key={i} className="flex items-center gap-2 bg-white/10 border border-white/20 rounded-full px-4 py-1.5">
                    <b.icon className="w-3.5 h-3.5 text-indigo-200" />
                    <span className="text-sm text-indigo-100 font-medium">{b.label}</span>
                  </div>
                ))}
              </div>
            </div>
            <div className="relative">
              <div className="absolute -inset-4 bg-white/5 rounded-3xl blur-xl" />
              <div className="relative z-10">
                <CTAMiniDashboard />
              </div>
            </div>
          </div>
        </div>
      </section>

    </main>
  );
}


