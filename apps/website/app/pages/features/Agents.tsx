'use client';

import { motion } from 'motion/react';
import Link from 'next/link';
import { MagicCard } from '../../components/ui/MagicCard';
import { FAQAccordion } from '../../components/ui/FAQAccordion';
import {
  Bot, Activity, Zap, GitBranch, RotateCcw, Shield,
  Eye, AlertTriangle, CheckCircle, XCircle, Clock, DollarSign,
  ArrowLeft, Network, Layers, Lock, CreditCard, Globe,
  Wrench, Users, Cpu, BookOpen, BarChart3,
  ChevronRight, TrendingDown, MessagesSquare,
  Search, Building, Share2, Terminal
} from 'lucide-react';
import { useState, useEffect } from 'react';

// ─── Hero Dashboard ───────────────────────────────────────────────────────────
function AgentsDashboard() {
  const [runs, setRuns] = useState(1240);
  const [successRate] = useState('۹۸.۷');

  useEffect(() => {
    const t = setInterval(() => setRuns(v => v + Math.floor(Math.random() * 2 + 1)), 2200);
    return () => clearInterval(t);
  }, []);

  const agentRuns = [
    { id: 'run-7821', agent: 'Support Agent', model: 'gpt-4o', steps: 6, latency: '2.1s', cost: '$0.012', status: 'موفق', ok: true, guardrail: 'Pass' },
    { id: 'run-7820', agent: 'Research Agent', model: 'claude-3', steps: 9, latency: '4.3s', cost: '$0.031', status: 'Retry', ok: null, guardrail: 'Warning' },
    { id: 'run-7819', agent: 'Data Agent', model: 'gpt-4o', steps: 4, latency: '1.8s', cost: '$0.008', status: 'موفق', ok: true, guardrail: 'Pass' },
    { id: 'run-7818', agent: 'Workflow Bot', model: 'gemini', steps: 12, latency: '6.7s', cost: '$0.044', status: 'خطا', ok: false, guardrail: 'Fail' },
  ];

  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl overflow-hidden">
      {/* Window bar */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-slate-200 dark:border-slate-800 bg-white/50 dark:bg-slate-900/50">
        <div className="w-3 h-3 rounded-full bg-rose-500/80" />
        <div className="w-3 h-3 rounded-full bg-amber-500/80" />
        <div className="w-3 h-3 rounded-full bg-emerald-500/80" />
        <div className="mx-auto flex items-center gap-2 px-3 py-1 rounded-md bg-slate-100 dark:bg-slate-800 text-xs text-slate-500 dark:text-slate-400 font-medium font-sans">
          <Bot className="w-3 h-3" />
          GodarAI AI Agents
        </div>
      </div>

      <div className="p-5 space-y-4">
        {/* Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {[
            { label: 'اجرای ایجنت امروز', value: runs.toLocaleString('fa-IR'), icon: Bot, color: 'text-indigo-500', bg: 'bg-indigo-50 dark:bg-indigo-500/10' },
            { label: 'نرخ موفقیت', value: `${successRate}٪`, icon: CheckCircle, color: 'text-emerald-500', bg: 'bg-emerald-50 dark:bg-emerald-500/10' },
            { label: 'میانگین تأخیر', value: '۲.۳s', icon: Zap, color: 'text-amber-500', bg: 'bg-amber-50 dark:bg-amber-500/10' },
            { label: 'کاهش هزینه', value: '۱۸٪', icon: TrendingDown, color: 'text-purple-500', bg: 'bg-purple-50 dark:bg-purple-500/10' },
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

        {/* Agent Runs Table + Trace mini */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          <div className="md:col-span-2 p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <div className="flex justify-between items-center mb-3">
              <h3 className="text-xs font-bold text-slate-900 dark:text-white flex items-center gap-1.5">
                <Activity className="w-3 h-3 text-indigo-500" /> Agent Runs
              </h3>
              <span className="flex items-center gap-1 text-[10px] text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-500/10 px-2 py-0.5 rounded-full font-medium">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse inline-block" />
                Live
              </span>
            </div>
            <div className="space-y-2 font-sans text-[10px]">
              {agentRuns.map((r, i) => (
                <div key={i} className="flex items-center gap-2 bg-slate-50 dark:bg-slate-800/40 rounded-lg px-3 py-2">
                  <Bot className="w-3 h-3 text-slate-400 shrink-0" />
                  <span className="text-slate-700 dark:text-slate-300 font-medium w-24 truncate">{r.agent}</span>
                  <span className="text-slate-400 font-sans w-16 truncate">{r.model}</span>
                  <span className="text-slate-500">{r.steps} steps</span>
                  <span className="text-slate-400 w-10">{r.latency}</span>
                  <span className="text-slate-400 w-12 font-sans">{r.cost}</span>
                  <span className={`px-1.5 py-0.5 rounded-full font-bold text-[9px] ml-auto ${
                    r.ok === true ? 'bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400' :
                    r.ok === false ? 'bg-rose-100 dark:bg-rose-500/20 text-rose-700 dark:text-rose-400' :
                    'bg-amber-100 dark:bg-amber-500/20 text-amber-700 dark:text-amber-400'
                  }`}>{r.status}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Guardrail Verdicts */}
          <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <h3 className="text-xs font-bold text-slate-900 dark:text-white mb-3 flex items-center gap-1.5">
              <Shield className="w-3 h-3 text-purple-500" /> Guardrail Verdict
            </h3>
            <div className="space-y-2 font-sans text-[10px]">
              {[
                { check: 'Prompt Injection', result: 'Blocked', color: 'text-rose-600 dark:text-rose-400', bg: 'bg-rose-50 dark:bg-rose-500/10' },
                { check: 'PII Detection', result: 'Warning', color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10' },
                { check: 'Tool Call Policy', result: 'Pass', color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10' },
                { check: 'Output Format', result: 'Fail', color: 'text-rose-600 dark:text-rose-400', bg: 'bg-rose-50 dark:bg-rose-500/10' },
                { check: 'Final Verdict', result: 'Retry', color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10' },
              ].map((v, i) => (
                <div key={i} className="flex items-center justify-between py-1 border-b border-slate-50 dark:border-slate-800/50 last:border-0">
                  <span className="text-slate-600 dark:text-slate-400">{v.check}</span>
                  <span className={`px-1.5 py-0.5 rounded-full font-bold text-[9px] ${v.bg} ${v.color}`}>{v.result}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Routing path */}
        <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
          <h3 className="text-xs font-bold text-slate-900 dark:text-white mb-3 flex items-center gap-1.5">
            <GitBranch className="w-3 h-3 text-indigo-500" /> تلاش دوباره و مسیر جایگزین
          </h3>
          <div className="flex flex-wrap items-center gap-2 font-sans text-[10px]">
            {[
              { label: 'Agent Request', color: 'bg-indigo-100 dark:bg-indigo-500/20 text-indigo-700 dark:text-indigo-400' },
              { label: '→', color: 'text-slate-400', plain: true },
              { label: 'Input Guardrail', color: 'bg-purple-100 dark:bg-purple-500/20 text-purple-700 dark:text-purple-400' },
              { label: '→', color: 'text-slate-400', plain: true },
              { label: 'Model A', color: 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300' },
              { label: '⚠ Fail →', color: 'text-amber-600 dark:text-amber-400', plain: true },
              { label: 'Retry', color: 'bg-amber-100 dark:bg-amber-500/20 text-amber-700 dark:text-amber-400' },
              { label: '/ Fallback →', color: 'text-slate-400', plain: true },
              { label: 'Model B', color: 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300' },
              { label: '→', color: 'text-slate-400', plain: true },
              { label: 'Output Guardrail', color: 'bg-purple-100 dark:bg-purple-500/20 text-purple-700 dark:text-purple-400' },
              { label: '→', color: 'text-slate-400', plain: true },
              { label: '✓ پاسخ نهایی', color: 'bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400' },
            ].map((item, i) => (
              item.plain
                ? <span key={i} className={`font-bold ${item.color}`}>{item.label}</span>
                : <span key={i} className={`px-2 py-0.5 rounded-full font-bold ${item.color}`}>{item.label}</span>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

// ─── Agent Trace Timeline ─────────────────────────────────────────────────────
function AgentTraceTimeline() {
  const steps = [
    { step: 1, label: 'User Prompt', type: 'prompt', latency: '—', cost: '—', icon: MessagesSquare, color: 'bg-indigo-500', ok: true },
    { step: 2, label: 'System Prompt', type: 'prompt', latency: '—', cost: '—', icon: Terminal, color: 'bg-slate-500', ok: true },
    { step: 3, label: 'Planning Step', type: 'planning', latency: '0.3s', cost: '$0.001', icon: Cpu, color: 'bg-purple-500', ok: true },
    { step: 4, label: 'Model Call', type: 'llm', latency: '1.2s', cost: '$0.008', icon: Bot, color: 'bg-indigo-600', ok: true },
    { step: 5, label: 'Tool Call', type: 'tool', latency: '0.8s', cost: '$0.000', icon: Wrench, color: 'bg-amber-500', ok: true },
    { step: 6, label: 'Observation', type: 'obs', latency: '0.1s', cost: '—', icon: Eye, color: 'bg-teal-500', ok: true },
    { step: 7, label: 'Retry / Fallback', type: 'retry', latency: '1.5s', cost: '$0.006', icon: RotateCcw, color: 'bg-amber-600', ok: null },
    { step: 8, label: 'Final Response', type: 'response', latency: '—', cost: '—', icon: CheckCircle, color: 'bg-emerald-500', ok: true },
  ];

  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs">
      <div className="flex items-center gap-2 mb-4">
        <div className="w-7 h-7 rounded-lg bg-indigo-100 dark:bg-indigo-500/20 flex items-center justify-center">
          <Activity className="w-3.5 h-3.5 text-indigo-600 dark:text-indigo-400" />
        </div>
        <span className="font-bold text-slate-800 dark:text-white">Agent Trace — run-7821</span>
        <span className="mr-auto text-[10px] font-bold text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-500/10 px-2 py-0.5 rounded-full">موفق</span>
      </div>
      <div className="space-y-0">
        {steps.map((s, i) => (
          <div key={i} className="flex gap-3">
            <div className="flex flex-col items-center">
              <div className={`w-6 h-6 rounded-full ${s.color} flex items-center justify-center shrink-0 z-10`}>
                <s.icon className="w-3 h-3 text-white" />
              </div>
              {i < steps.length - 1 && (
                <div className="w-px flex-1 bg-slate-200 dark:bg-slate-700 my-0.5" style={{ minHeight: '20px' }} />
              )}
            </div>
            <div className="pb-3 flex-1">
              <div className="flex items-center gap-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg px-3 py-2">
                <span className="text-[9px] font-bold text-slate-400 w-4">{s.step}</span>
                <span className="font-medium text-slate-700 dark:text-slate-300 flex-1">{s.label}</span>
                <span className={`px-1.5 py-0.5 rounded text-[8px] font-bold font-sans ${
                  s.type === 'tool' ? 'bg-amber-100 dark:bg-amber-500/20 text-amber-700 dark:text-amber-400' :
                  s.type === 'llm' ? 'bg-indigo-100 dark:bg-indigo-500/20 text-indigo-700 dark:text-indigo-400' :
                  s.type === 'retry' ? 'bg-amber-100 dark:bg-amber-500/20 text-amber-700 dark:text-amber-400' :
                  'bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400'
                }`}>{s.type}</span>
                {s.latency !== '—' && <span className="text-slate-400 text-[9px] font-sans">{s.latency}</span>}
                {s.cost !== '—' && <span className="text-slate-400 text-[9px] font-sans">{s.cost}</span>}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Metrics Dashboard ────────────────────────────────────────────────────────
function AgentMetricsDashboard() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs space-y-3">
      <div className="flex items-center gap-2 mb-1">
        <div className="w-7 h-7 rounded-lg bg-amber-100 dark:bg-amber-500/20 flex items-center justify-center">
          <BarChart3 className="w-3.5 h-3.5 text-amber-600 dark:text-amber-400" />
        </div>
        <span className="font-bold text-slate-800 dark:text-white">سنجه‌های ایجنت — امروز</span>
      </div>
      <div className="grid grid-cols-2 gap-2">
        {[
          { label: 'نرخ موفقیت ایجنت', value: '۹۸.۷٪', bar: 99, color: 'bg-emerald-500', tc: 'text-emerald-600 dark:text-emerald-400' },
          { label: 'میانگین تأخیر', value: '۲.۳s', bar: 45, color: 'bg-amber-500', tc: 'text-amber-600 dark:text-amber-400' },
          { label: 'مصرف توکن', value: '۴.۲M', bar: 70, color: 'bg-indigo-500', tc: 'text-indigo-600 dark:text-indigo-400' },
          { label: 'هزینه امروز', value: '$۱۸.۴', bar: 38, color: 'bg-purple-500', tc: 'text-purple-600 dark:text-purple-400' },
          { label: 'Error Rate', value: '۱.۳٪', bar: 13, color: 'bg-rose-500', tc: 'text-rose-600 dark:text-rose-400' },
          { label: 'زمان فراخوانی ابزار', value: '۰.۹s', bar: 30, color: 'bg-teal-500', tc: 'text-teal-600 dark:text-teal-400' },
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
          { label: 'هزینه بر اساس ایجنت', vals: [{ name: 'Support', pct: '45٪' }, { name: 'Research', pct: '31٪' }, { name: 'Data', pct: '24٪' }] },
          { label: 'هزینه بر اساس مدل', vals: [{ name: 'gpt-4o', pct: '62٪' }, { name: 'claude-3', pct: '28٪' }, { name: 'gemini', pct: '10٪' }] },
        ].map((g, i) => (
          <div key={i} className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-2.5">
            <div className="font-bold text-slate-600 dark:text-slate-400 mb-1.5">{g.label}</div>
            {g.vals.map((v, j) => (
              <div key={j} className="flex justify-between py-0.5 border-b border-slate-50 dark:border-slate-800/50 last:border-0">
                <span className="text-slate-500">{v.name}</span>
                <span className="font-bold text-slate-700 dark:text-slate-300">{v.pct}</span>
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Agent Guardrails Panel ───────────────────────────────────────────────────
function AgentGuardrailPanel() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs space-y-3">
      <div className="flex items-center gap-2">
        <div className="w-7 h-7 rounded-lg bg-purple-100 dark:bg-purple-500/20 flex items-center justify-center">
          <Shield className="w-3.5 h-3.5 text-purple-600 dark:text-purple-400" />
        </div>
        <span className="font-bold text-slate-800 dark:text-white">Guardrails — Agent Output Check</span>
      </div>
      <div className="space-y-2">
        {[
          { name: 'Prompt Injection', result: 'Blocked', ok: false, note: 'ورودی مخرب شناسایی شد' },
          { name: 'PII Detection', result: 'Warning', ok: null, note: 'شماره تماس در پاسخ' },
          { name: 'Tool Call Policy', result: 'Pass', ok: true, note: '' },
          { name: 'Output Format', result: 'Fail', ok: false, note: 'فرمت JSON نامعتبر' },
        ].map((c, i) => (
          <div key={i} className="flex items-center gap-2 px-3 py-2 rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <div className={`w-4 h-4 rounded-full flex items-center justify-center shrink-0 ${
              c.ok === true ? 'bg-emerald-100 dark:bg-emerald-500/20' :
              c.ok === false ? 'bg-rose-100 dark:bg-rose-500/20' :
              'bg-amber-100 dark:bg-amber-500/20'
            }`}>
              {c.ok === true && <CheckCircle className="w-2.5 h-2.5 text-emerald-600 dark:text-emerald-400" />}
              {c.ok === false && <XCircle className="w-2.5 h-2.5 text-rose-600 dark:text-rose-400" />}
              {c.ok === null && <AlertTriangle className="w-2.5 h-2.5 text-amber-600 dark:text-amber-400" />}
            </div>
            <span className="flex-1 text-slate-700 dark:text-slate-300">{c.name}</span>
            {c.note && <span className="text-[9px] text-slate-400 hidden sm:block">{c.note}</span>}
            <span className={`px-1.5 py-0.5 rounded-full font-bold text-[9px] ${
              c.result === 'Pass' ? 'bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400' :
              c.result === 'Blocked' || c.result === 'Fail' ? 'bg-rose-100 dark:bg-rose-500/20 text-rose-700 dark:text-rose-400' :
              'bg-amber-100 dark:bg-amber-500/20 text-amber-700 dark:text-amber-400'
            }`}>{c.result}</span>
          </div>
        ))}
      </div>
      <div className="flex items-center justify-between p-3 rounded-xl bg-indigo-50 dark:bg-indigo-500/10 border border-indigo-200 dark:border-indigo-500/30">
        <span className="text-indigo-700 dark:text-indigo-400 font-bold">Final Verdict</span>
        <span className="text-indigo-700 dark:text-indigo-300 font-bold text-[11px] bg-indigo-100 dark:bg-indigo-500/20 px-2 py-0.5 rounded-full">Retry</span>
      </div>
    </div>
  );
}

// ─── Gateway Diagram ──────────────────────────────────────────────────────────
function GatewayDiagram() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs">
      <div className="flex items-center gap-2 mb-4">
        <div className="w-7 h-7 rounded-lg bg-emerald-100 dark:bg-emerald-500/20 flex items-center justify-center">
          <Network className="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
        </div>
        <span className="font-bold text-slate-800 dark:text-white">یک درگاه — چند مدل</span>
      </div>
      <div className="flex flex-col items-center gap-3">
        {/* Agent App */}
        <div className="px-4 py-2 rounded-xl bg-indigo-100 dark:bg-indigo-500/20 border border-indigo-200 dark:border-indigo-500/30 text-indigo-700 dark:text-indigo-300 font-bold flex items-center gap-2">
          <Bot className="w-3.5 h-3.5" /> Agent App
        </div>
        <ChevronRight className="w-4 h-4 text-slate-400 rotate-90" />
        {/* Gateway */}
        <div className="px-6 py-3 rounded-xl bg-slate-900 dark:bg-slate-950 border border-indigo-500/40 text-white font-bold flex items-center gap-2 shadow-[0_0_20px_rgba(99,102,241,0.15)]">
          <Shield className="w-3.5 h-3.5 text-indigo-400" /> GodarAI Gateway
        </div>
        <div className="flex items-center gap-2">
          <div className="w-px h-4 bg-slate-300 dark:bg-slate-700" />
        </div>
        {/* Models */}
        <div className="grid grid-cols-3 gap-2 w-full">
          {[
            { name: 'Model A', sub: 'gpt-4o', color: 'bg-indigo-50 dark:bg-indigo-500/10 text-indigo-700 dark:text-indigo-400 border-indigo-200 dark:border-indigo-500/30' },
            { name: 'Model B', sub: 'claude-3', color: 'bg-purple-50 dark:bg-purple-500/10 text-purple-700 dark:text-purple-400 border-purple-200 dark:border-purple-500/30' },
            { name: 'Model C', sub: 'gemini', color: 'bg-teal-50 dark:bg-teal-500/10 text-teal-700 dark:text-teal-400 border-teal-200 dark:border-teal-500/30' },
          ].map((m, i) => (
            <div key={i} className={`p-2 rounded-lg border text-center ${m.color}`}>
              <div className="font-bold text-[10px]">{m.name}</div>
              <div className="text-[9px] opacity-70">{m.sub}</div>
            </div>
          ))}
        </div>
        {/* Services */}
        <div className="grid grid-cols-3 gap-2 w-full mt-1">
          {[
            { label: 'Tool Services', icon: Wrench, color: 'text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-500/10 border-amber-200 dark:border-amber-500/30' },
            { label: 'Guardrails', icon: Shield, color: 'text-rose-600 dark:text-rose-400 bg-rose-50 dark:bg-rose-500/10 border-rose-200 dark:border-rose-500/30' },
            { label: 'Observability', icon: Eye, color: 'text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-500/10 border-emerald-200 dark:border-emerald-500/30' },
          ].map((s, i) => (
            <div key={i} className={`p-2 rounded-lg border flex flex-col items-center gap-1 ${s.color}`}>
              <s.icon className="w-3 h-3" />
              <span className="text-[9px] font-medium text-center leading-tight">{s.label}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── Tool Call Table ──────────────────────────────────────────────────────────
function ToolCallTable() {
  return (
    <div className="bg-slate-900 dark:bg-slate-950 border border-slate-800 rounded-2xl p-5 font-sans text-xs">
      <div className="flex items-center justify-between mb-4">
        <span className="text-sm font-bold text-white flex items-center gap-2">
          <Wrench className="w-4 h-4 text-amber-400" /> Tool Calls — Debug View
        </span>
        <span className="text-[10px] text-slate-400 bg-slate-800 px-2 py-0.5 rounded-full">run-7821</span>
      </div>
      <table className="w-full text-right">
        <thead>
          <tr className="text-slate-500 border-b border-slate-800">
            {['نام ابزار', 'ورودی', 'خروجی', 'وضعیت', 'تأخیر', 'مرحله'].map(h => (
              <th key={h} className="pb-2 font-medium text-[10px]">{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {[
            { tool: 'search_web', input: '{ query: "..." }', output: '[results]', status: 'Pass', latency: '0.8s', step: 5, ok: true },
            { tool: 'read_file', input: '{ path: "..." }', output: 'null', status: 'Error', latency: '0.2s', step: 7, ok: false },
            { tool: 'send_email', input: '{ to: "..." }', output: 'sent', status: 'Pass', latency: '1.1s', step: 9, ok: true },
            { tool: 'calc_metric', input: '{ data: [...] }', output: '{ score: 0.87 }', status: 'Pass', latency: '0.3s', step: 11, ok: true },
          ].map((r, i) => (
            <tr key={i} className="border-b border-slate-800/50 last:border-0">
              <td className="py-2 text-amber-400 font-medium font-sans">{r.tool}</td>
              <td className="py-2 text-slate-400 font-sans truncate max-w-16">{r.input}</td>
              <td className="py-2 text-slate-300 font-sans truncate max-w-16">{r.output}</td>
              <td className="py-2">
                <span className={`px-1.5 py-0.5 rounded-full text-[9px] font-bold ${r.ok ? 'bg-emerald-500/20 text-emerald-400' : 'bg-rose-500/20 text-rose-400'}`}>{r.status}</span>
              </td>
              <td className="py-2 text-slate-400">{r.latency}</td>
              <td className="py-2 text-slate-500">{r.step}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ─── Three-Layer Architecture ─────────────────────────────────────────────────
function ThreeLayerArch() {
  return (
    <div className="space-y-2 font-sans text-sm">
      {[
        { label: 'Gateway Layer', desc: 'Routing · Fallback · Load Balance · Virtual Keys', icon: Network, color: 'bg-indigo-500', light: 'bg-indigo-50 dark:bg-indigo-500/10 border-indigo-200 dark:border-indigo-500/30 text-indigo-700 dark:text-indigo-300' },
        { label: 'Guardrails Layer', desc: 'Input Check · Output Check · Policy · Block / Fix / Retry', icon: Shield, color: 'bg-purple-500', light: 'bg-purple-50 dark:bg-purple-500/10 border-purple-200 dark:border-purple-500/30 text-purple-700 dark:text-purple-300' },
        { label: 'Observability Layer', desc: 'Trace · Metrics · Logs · Debug · Alerts', icon: Eye, color: 'bg-emerald-500', light: 'bg-emerald-50 dark:bg-emerald-500/10 border-emerald-200 dark:border-emerald-500/30 text-emerald-700 dark:text-emerald-300' },
      ].map((layer, i) => (
        <div key={i} className={`flex items-center gap-3 p-3 rounded-xl border ${layer.light}`}>
          <div className={`w-8 h-8 rounded-lg ${layer.color} flex items-center justify-center shrink-0`}>
            <layer.icon className="w-4 h-4 text-white" />
          </div>
          <div>
            <div className="font-bold text-[12px]">{layer.label}</div>
            <div className="text-[10px] opacity-70">{layer.desc}</div>
          </div>
        </div>
      ))}
      <div className="mt-3 text-center text-[10px] text-slate-400 border-t border-slate-200 dark:border-slate-700 pt-3">
        همه Agentها از این سه لایه عبور می‌کنند
      </div>
    </div>
  );
}

// ─── CTA Mini Dashboard ───────────────────────────────────────────────────────
function CTAMiniDashboard() {
  const [runs, setRuns] = useState(1240);
  useEffect(() => {
    const t = setInterval(() => setRuns(v => v + Math.floor(Math.random() * 2 + 1)), 2200);
    return () => clearInterval(t);
  }, []);
  return (
    <div className="bg-white/10 backdrop-blur-sm border border-white/20 rounded-2xl p-5 font-sans text-sm text-white">
      <div className="flex items-center gap-2 mb-4 pb-3 border-b border-white/10">
        <Bot className="w-4 h-4 text-indigo-300" />
        <span className="text-xs font-bold text-white/80">ایجنت‌های هوش مصنوعی — نمای نمونه</span>
        <span className="mr-auto flex items-center gap-1 text-[10px] text-emerald-300 bg-emerald-500/20 px-2 py-0.5 rounded-full">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse inline-block" />
          Live
        </span>
      </div>
      <div className="grid grid-cols-2 gap-3">
        {[
          { label: `${runs.toLocaleString('fa-IR')} اجرای ایجنت امروز`, icon: Bot, color: 'text-indigo-300' },
          { label: '۹۸.۷٪ نرخ موفقیت', icon: CheckCircle, color: 'text-emerald-300' },
          { label: 'سنجه‌های مصرف و عملکرد', icon: BarChart3, color: 'text-amber-300' },
          { label: '۲.۳s میانگین تأخیر', icon: Zap, color: 'text-teal-300' },
          { label: 'مسیریابی بین مقصدهای مدل', icon: TrendingDown, color: 'text-purple-300' },
          { label: 'Retry / Fallback / Route / Block', icon: GitBranch, color: 'text-rose-300' },
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

// FAQ
const faqItems = [
  {
    question: 'گدارAI در اجرای ایجنت چه نقشی دارد؟',
    answer: 'گدارAI لایه اجرای خود ایجنت نیست؛ فراخوانی‌های مدل را از یک درگاه قابل کنترل عبور می‌دهد تا دسترسی، مسیریابی، گاردریل، مصرف، لاگ و رد درخواست متمرکز بمانند.',
  },
  {
    question: 'آیا می‌توانم مسیر فراخوانی مدل را مرحله‌به‌مرحله ببینم؟',
    answer: 'بله. رد درخواست تلاش‌ها، مقصدهای مدل، خطا، تلاش دوباره و مسیر جایگزین را نشان می‌دهد. برای مراحل داخلی و فراخوانی ابزار ایجنت، داده ردیابی چارچوب خود را نیز با OpenTelemetry همبسته کنید.',
  },
  {
    question: 'اگر فراخوانی مدل خطا بدهد یا کند باشد چه می‌شود؟',
    answer: 'در مدل مجازی می‌توانید مهلت پاسخ، کدهای تلاش دوباره و مقصدهای جایگزین را تنظیم کنید. نتیجه هر تلاش در رد درخواست قابل پیگیری است.',
  },
  {
    question: 'آیا می‌توانم روی متن ورودی و خروجی ایجنت گاردریل اجرا کنم؟',
    answer: 'بله. در فراخوانی `Chat Completions` می‌توانید گاردریل ورودی یا خروجی را انتخاب کنید تا متن بررسی، نشانه‌گذاری، متوقف یا در نوع‌های پشتیبانی‌شده اصلاح شود.',
  },
  {
    question: 'چطور هزینه Agentها را کنترل کنم؟',
    answer: 'مصرف توکن و هزینه فراخوانی‌های مدل را بر اساس تیم یا حساب مجازی بررسی کنید و با سقف هزینه، محدودیت نرخ، حافظه نهان و انتخاب مدل مناسب کنترل کنید.',
  },
  {
    question: 'کدام ایجنت‌ها می‌توانند از این لایه استفاده کنند؟',
    answer: 'هر ایجنتی که فراخوانی مدل خود را از API سازگار گدارAI بفرستد؛ از دستیار محصول و پشتیبانی تا جریان پژوهش یا اتوماسیون داخلی.',
  },
  {
    question: 'آیا برای استفاده از گدارAI نیاز به پرداخت ارزی دارم؟',
    answer: 'خیر. گدارAI برای تیم‌های ایرانی طراحی شده و هدف آن ساده‌کردن دسترسی به مدل‌های AI و اجرای Agentها بدون نیاز به پرداخت ارزی مستقیم و با ریسک کمتر نسبت به اتصال مستقیم به سرویس‌های خارجی است.',
  },
];

// ─── Main Page ────────────────────────────────────────────────────────────────
export function Agents() {
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
              <Bot className="w-4 h-4" />
              اجرای ایجنت از مسیر گدارAI
            </motion.div>

            <motion.h1
              className="text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight mb-8 leading-tight text-slate-900 dark:text-white"
              style={{ fontFamily: 'var(--font-heading)' }}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
            >
              فراخوانی‌های مدل ایجنت را از{' '}
              <span className="text-transparent bg-clip-text bg-gradient-to-l from-indigo-500 to-purple-500">
                نمونه آزمایشی تا محیط عملیاتی
              </span>{' '}
              قابل کنترل نگه دارید
            </motion.h1>

            <motion.p
              className="text-lg md:text-xl text-slate-600 dark:text-slate-400 mb-10 leading-relaxed max-w-3xl mx-auto"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
            >
              فراخوانی‌های مدل ایجنت را از درگاه گدارAI بفرستید تا مدل مقصد، دسترسی، گاردریل، هزینه و مسیر هر درخواست در یک نقطه مدیریت و بررسی شود.
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
            <AgentsDashboard />
          </motion.div>
        </div>
      </section>

      {/* ═══ 2. Trust Bar ════════════════════════════════════════════════════ */}
      <section className="py-12 border-y border-slate-100 dark:border-slate-800/50 bg-slate-50/50 dark:bg-slate-900/20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-8">
            <h2 className="text-xl font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
              برای تیم‌هایی که ایجنت را در محصول واقعی اجرا می‌کنند
            </h2>
            <p className="text-sm text-slate-600 dark:text-slate-400 max-w-2xl mx-auto">
              گدارAI مسئولیت اتصال ایجنت به مدل را از منطق خود ایجنت جدا می‌کند. چارچوب ایجنت همچنان گردش کار و ابزارها را اجرا می‌کند و درگاه، فراخوانی مدل را کنترل و پایش می‌کند.
            </p>
          </div>
          <div className="flex flex-wrap justify-center gap-3">
            {[
              'استارتاپ‌های هوش مصنوعی', 'تیم‌های بک‌اند و زیرساخت', 'محصولات SaaS',
              'تیم‌های پشتیبانی مشتری', 'سازمان‌های داده‌محور', 'تیم‌های تحقیق و توسعه',
              'تیم‌های دارای جریان چندمرحله‌ای',
              'شرکت‌هایی که ایجنت داخلی یا محصولی می‌سازند',
            ].map((chip, idx) => (
              <div key={idx} className="px-4 py-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-full text-sm font-medium text-slate-700 dark:text-slate-300 shadow-sm">
                {chip}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 3. Deploy to Production ══════════════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              کنترل‌های فراخوانی مدل را به ایجنت اضافه کنید
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              درگاه، گاردریل و مشاهده‌پذیری گدارAI روی فراخوانی مدل کار می‌کنند. منطق برنامه‌ریزی، حافظه و اجرای ابزار همچنان در چارچوب ایجنت شما باقی می‌ماند.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Card 1 — Trace */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(99,102,241,0.08)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-indigo-100 dark:bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 rounded-xl flex items-center justify-center mb-6">
                  <Activity className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  رد درخواست برای هر فراخوانی مدل
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  تلاش‌ها، مدل مقصد، خطا، تلاش دوباره و مسیر جایگزین هر درخواست را ببینید. برای مراحل داخلی و ابزارها، ردیابی چارچوب ایجنت را در کنار این داده‌ها نگه دارید.
                </p>
                <p className="text-xs text-indigo-600 dark:text-indigo-400 font-medium mb-6">
                  برای عیب‌یابی اتصال مدل در جریان‌های چندمرحله‌ای.
                </p>
                <AgentTraceTimeline />
              </div>
            </MagicCard>

            {/* Card 2 — Metrics */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(245,158,11,0.07)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-amber-100 dark:bg-amber-500/20 text-amber-600 dark:text-amber-400 rounded-xl flex items-center justify-center mb-6">
                  <BarChart3 className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  پایش عملکرد و هزینه مدل
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  تعداد درخواست، تأخیر، توکن، هزینه و نرخ خطا را بر اساس مدل، ارائه‌دهنده، تیم یا حساب مجازی بررسی کنید و مصرف سرویس ایجنت را از دیگر سرویس‌ها جدا نگه دارید.
                </p>
                <p className="text-xs text-amber-600 dark:text-amber-400 font-medium mb-6">
                  برای هر ایجنت یا سرویس، هویت و VAT مستقل در نظر بگیرید.
                </p>
                <AgentMetricsDashboard />
              </div>
            </MagicCard>

            {/* Card 3 — Guardrails */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(139,92,246,0.07)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-purple-100 dark:bg-purple-500/20 text-purple-600 dark:text-purple-400 rounded-xl flex items-center justify-center mb-6">
                  <Shield className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  کنترل متن با گاردریل
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  روی متن ورودی و پاسخ کامل مدل، گاردریل‌هایی مانند عبارت مسدود، PII پایه، تشخیص Secret یا اعتبارسنجی JSON اجرا کنید.
                </p>
                <p className="text-xs text-purple-600 dark:text-purple-400 font-medium mb-6">
                  گاردریل متن جای کنترل مجوز ابزار یا منطق خود ایجنت را نمی‌گیرد.
                </p>
                <AgentGuardrailPanel />
              </div>
            </MagicCard>

            {/* Card 4 — Gateway */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(16,185,129,0.07)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-emerald-100 dark:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 rounded-xl flex items-center justify-center mb-6">
                  <Network className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  اتصال ایجنت به چند مدل با یک درگاه
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  شناسه مدل یا مدل مجازی را در فراخوانی بفرستید. در مدل مجازی می‌توانید چند مقصد سازگار، تلاش دوباره و مسیر جایگزین تعریف کنید.
                </p>
                <p className="text-xs text-emerald-600 dark:text-emerald-400 font-medium mb-6">
                  شناسه مصرف ایجنت را ثابت و جزئیات مقصد را در درگاه نگه دارید.
                </p>
                <GatewayDiagram />
              </div>
            </MagicCard>
          </div>
        </div>
      </section>

      {/* ═══ 4. Agent Observability ═══════════════════════════════════════════ */}
      <section className="py-24 bg-slate-900 dark:bg-[#060913] text-white border-y border-slate-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              مشاهده‌پذیری فراخوانی مدل، در کنار ردیابی خود ایجنت
            </h2>
            <p className="text-lg text-slate-400">
              ایجنت فقط یک فراخوانی مدل نیست. رد درخواست گدارAI مسیر داخل درگاه را نشان می‌دهد؛ برای تصمیم‌ها و ابزارهای ایجنت، ردیابی چارچوب یا اپلیکیشن خود را نیز نگه دارید.
            </p>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            {/* Left: Capabilities */}
            <div className="space-y-6">
              {[
                {
                  icon: Activity, color: 'text-indigo-400', bg: 'bg-indigo-500/20',
                  title: 'رد درخواست درگاه',
                  desc: 'مدل مقصد، تلاش‌ها، خطا، تلاش دوباره، مسیر جایگزین و زمان مراحل داخلی درگاه را بررسی کنید.',
                },
                {
                  icon: Wrench, color: 'text-amber-400', bg: 'bg-amber-500/20',
                  title: 'همبستگی با ردیابی اپلیکیشن',
                  desc: 'با `request_id` و `trace_id`، فراخوانی مدل را به مرحله مربوط در ردیابی ایجنت یا سرویس خود وصل کنید.',
                },
                {
                  icon: BarChart3, color: 'text-emerald-400', bg: 'bg-emerald-500/20',
                  title: 'مقایسه مدل و نسخه پرامپت',
                  desc: 'مدل، شناسه نسخه پرامپت، هزینه و نتیجه درخواست را کنار ارزیابی کیفیت محصول خود مقایسه کنید.',
                },
                {
                  icon: Users, color: 'text-purple-400', bg: 'bg-purple-500/20',
                  title: 'پایش بر اساس هویت سرویس',
                  desc: 'برای هر ایجنت VAT جدا بسازید تا مصرف، هزینه و خطای فراخوانی‌های آن از دیگر سرویس‌ها تفکیک شود.',
                },
              ].map((item, i) => (
                <div key={i} className="flex gap-4">
                  <div className={`w-10 h-10 rounded-xl ${item.bg} ${item.color} flex items-center justify-center shrink-0 mt-1`}>
                    <item.icon className="w-5 h-5" />
                  </div>
                  <div>
                    <h4 className="font-bold text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>{item.title}</h4>
                    <p className="text-slate-400 text-sm leading-relaxed">{item.desc}</p>
                  </div>
                </div>
              ))}
            </div>

            {/* Right: Tool Call Table */}
            <div className="relative">
              <div className="absolute -inset-4 bg-gradient-to-br from-indigo-500/10 to-amber-500/10 rounded-3xl blur-2xl z-0" />
              <div className="relative z-10">
                <ToolCallTable />
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ═══ 5. Production Controls ═══════════════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              کنترل‌های محیط عملیاتی برای فراخوانی مدل
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              وقتی ایجنت به داده و ابزار دسترسی دارد، کنترل مدل فقط بخشی از ایمنی است. مجوز ابزار و داده باید در خود اپلیکیشن یا لایه ابزار نیز اعمال شود.
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            {[
              {
                icon: Wrench, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10',
                title: 'هویت مستقل برای هر ایجنت',
                desc: 'برای هر سرویس یا ایجنت یک کاربر مجازی و VAT جدا بسازید تا دسترسی، ابطال و گزارش مصرف آن مستقل باشد.',
              },
              {
                icon: Clock, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10',
                title: 'مهلت پاسخ و تلاش دوباره مدل',
                desc: 'در مدل مجازی، مهلت پاسخ و کدهای تلاش دوباره را برای فراخوانی مدل تنظیم کنید. رفتار ابزارها در اختیار چارچوب ایجنت است.',
              },
              {
                icon: GitBranch, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10',
                title: 'مسیر جایگزین بین مدل‌ها',
                desc: 'برای خطاهای مشخص، مقصد بعدی مدل مجازی را امتحان کنید. کیفیت نامناسب پاسخ به ارزیابی و تصمیم اپلیکیشن نیاز دارد.',
              },
              {
                icon: Eye, color: 'text-teal-600 dark:text-teal-400', bg: 'bg-teal-50 dark:bg-teal-500/10',
                title: 'بررسی انسانی در منطق محصول',
                desc: 'برای تصمیم‌های پرریسک، خطا یا نتیجه گاردریل را در اپلیکیشن دریافت کنید و جریان بررسی انسانی خود را اجرا کنید.',
              },
              {
                icon: DollarSign, color: 'text-rose-600 dark:text-rose-400', bg: 'bg-rose-50 dark:bg-rose-500/10',
                title: 'سقف هزینه و محدودیت نرخ',
                desc: 'برای هویت سرویس یا تیم، سیاست مصرف تعریف کنید تا تعداد درخواست و هزینه فراخوانی مدل قابل کنترل باشد.',
              },
              {
                icon: BookOpen, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10',
                title: 'لاگ درخواست با حریم خصوصی',
                desc: 'ثبت محتوای درخواست و پاسخ را متناسب با داده ایجنت تنظیم کنید و مجوز مشاهده محتوا را محدود نگه دارید.',
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

      {/* ═══ 6. Framework-friendly ════════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              اتصال مستقل از چارچوب ایجنت
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              اگر چارچوب شما امکان تغییر نشانی پایه و توکن API سازگار را دارد، فراخوانی مدل را از گدارAI بفرستید و منطق ایجنت را در همان چارچوب نگه دارید.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {/* Framework Compatibility */}
            <MagicCard className="bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden" gradientColor="rgba(99,102,241,0.07)">
              <div className="p-6">
                <div className="w-10 h-10 bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 rounded-xl flex items-center justify-center mb-4">
                  <Layers className="w-5 h-5" />
                </div>
                <h3 className="font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                  سازگار با کلاینت‌های API
                </h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed mb-5">
                  در چارچوب‌های دارای کلاینت سازگار با OpenAI، نشانی پایه، توکن گدارAI و شناسه مدل را تنظیم و مسیر را در محیط خود آزمایش کنید.
                </p>
                <div className="grid grid-cols-2 gap-2 font-sans text-xs">
                  {['Agent Framework', 'Workflow Engine', 'Tool Calling SDK', 'Multi-agent System', 'Internal Agent Runtime', 'Custom Python Agent'].map(f => (
                    <div key={f} className="flex items-center gap-2 p-2 rounded-lg bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-700">
                      <Bot className="w-3 h-3 text-indigo-400 shrink-0" />
                      <span className="text-slate-600 dark:text-slate-400 text-[10px]">{f}</span>
                    </div>
                  ))}
                </div>
              </div>
            </MagicCard>

            {/* Three-Layer */}
            <MagicCard className="bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden" gradientColor="rgba(139,92,246,0.07)">
              <div className="p-6">
                <div className="w-10 h-10 bg-purple-50 dark:bg-purple-500/10 text-purple-600 dark:text-purple-400 rounded-xl flex items-center justify-center mb-4">
                  <Share2 className="w-5 h-5" />
                </div>
                <h3 className="font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                  درگاه، گاردریل و مشاهده‌پذیری در یک مسیر
                </h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed mb-5">
                  فراخوانی‌های مدل را از یک مسیر بفرستید تا دسترسی، گاردریل متن، لاگ، رد درخواست و هزینه با قرارداد مشترک مدیریت شوند.
                </p>
                <ThreeLayerArch />
              </div>
            </MagicCard>

            {/* Internal & Product */}
            <MagicCard className="bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden" gradientColor="rgba(16,185,129,0.06)">
              <div className="p-6">
                <div className="w-10 h-10 bg-emerald-50 dark:bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 rounded-xl flex items-center justify-center mb-4">
                  <Building className="w-5 h-5" />
                </div>
                <h3 className="font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                  مناسب برای ایجنت‌های داخلی و محصولی
                </h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed mb-5">
                  چه ایجنت داخلی بسازید و چه دستیار محصول، می‌توانید فراخوانی مدل آن را با هویت مستقل، مدل‌های مجاز و سیاست مصرف مشخص اجرا کنید.
                </p>
                <div className="grid grid-cols-2 gap-3 font-sans text-xs">
                  <div className="p-3 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-700">
                    <div className="font-bold text-slate-600 dark:text-slate-400 mb-2 text-[10px]">Internal Agents</div>
                    {['Support Assistant', 'Data Analysis Agent', 'Ops Agent'].map(a => (
                      <div key={a} className="flex items-center gap-1.5 py-1">
                        <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 shrink-0" />
                        <span className="text-slate-600 dark:text-slate-400 text-[10px]">{a}</span>
                      </div>
                    ))}
                  </div>
                  <div className="p-3 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-700">
                    <div className="font-bold text-slate-600 dark:text-slate-400 mb-2 text-[10px]">Product Agents</div>
                    {['Customer Assistant', 'Workflow Automation', 'Research Agent'].map(a => (
                      <div key={a} className="flex items-center gap-1.5 py-1">
                        <div className="w-1.5 h-1.5 rounded-full bg-indigo-400 shrink-0" />
                        <span className="text-slate-600 dark:text-slate-400 text-[10px]">{a}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </MagicCard>
          </div>
        </div>
      </section>

      {/* ═══ 7. Metrics ══════════════════════════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              داده‌های لازم برای نگهداری ایجنت در محیط عملیاتی
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              در محیط عملیاتی باید بتوانید اتصال مدل را عیب‌یابی کنید، هزینه را به هویت سرویس نسبت دهید و رفتار مسیر جایگزین را بسنجید.
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {[
              {
                icon: Activity, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10',
                value: 'هر', unit: 'درخواست مدل',
                label: 'رد درخواست',
                desc: 'مقصد، تلاش‌ها، خطا و زمان اجرای درخواست در درگاه را ببینید.',
              },
              {
                icon: BarChart3, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10',
                value: 'چند', unit: 'سنجه اصلی',
                label: 'روند مصرف و عملکرد',
                desc: 'تعداد درخواست، تأخیر، توکن، هزینه، خطا و اثربخشی حافظه نهان را بررسی کنید.',
              },
              {
                icon: Network, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10',
                value: 'چند', unit: 'مقصد',
                label: 'مدل مجازی',
                desc: 'چند مقصد سازگار را پشت یک شناسه پایدار برای ایجنت قرار دهید.',
              },
              {
                icon: GitBranch, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10',
                value: 'مستقل', unit: 'از کاربر',
                label: 'هویت سرویس با VAT',
                desc: 'توکن ایجنت را مستقل بسازید تا دسترسی، ابطال و پایش مصرف آن جدا باشد.',
              },
            ].map((m, i) => (
              <motion.div
                key={i}
                className="bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 text-center hover:border-indigo-200 dark:hover:border-indigo-800 transition-colors group"
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: i * 0.1 }}
              >
                <div className={`w-12 h-12 ${m.bg} ${m.color} rounded-xl flex items-center justify-center mx-auto mb-4 group-hover:scale-110 transition-transform`}>
                  <m.icon className="w-6 h-6" />
                </div>
                <div className="text-3xl font-bold text-slate-900 dark:text-white mb-1 font-sans">
                  {m.value}<span className="text-base font-medium text-slate-400 mr-1">{m.unit}</span>
                </div>
                <div className="font-bold text-slate-700 dark:text-slate-300 mb-3" style={{ fontFamily: 'var(--font-heading)' }}>{m.label}</div>
                <p className="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">{m.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 8. Differentiation ══════════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              چرا اجرای ایجنت از گدارAI برای تیم‌های ایرانی مهم است؟
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              برای تیم‌های ایرانی، عملیاتی کردن Agent فقط مسئله فنی نیست؛ دسترسی پایدار به مدل‌ها، پرداخت، تحریم، هزینه، داده فارسی و کنترل ریسک هم اهمیت دارد.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {[
              {
                icon: Shield, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10', border: 'border-indigo-100 dark:border-indigo-500/20',
                title: 'کاهش ریسک محدودیت‌های اتصال مستقیم',
                desc: 'درگاه مشترک، وابستگی کد ایجنت به اتصال مستقیم هر ارائه‌دهنده را کمتر می‌کند و تغییر مدل مقصد را ساده‌تر می‌سازد.',
              },
              {
                icon: CreditCard, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10', border: 'border-emerald-100 dark:border-emerald-500/20',
                title: 'بدون پرداخت ارزی مستقیم',
                desc: 'به‌جای درگیری با کارت اعتباری خارجی، پرداخت دلاری و حساب‌های بین‌المللی، مصرف Agentها و مدل‌ها را از طریق گدارAI و متناسب با نیاز بازار ایران مدیریت کنید.',
              },
              {
                icon: Lock, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10', border: 'border-purple-100 dark:border-purple-500/20',
                title: 'کنترل متمرکز مدل و هزینه',
                desc: 'اتصال مدل، مشاهده‌پذیری، گاردریل متن، مسیریابی و هزینه را در یک مسیر نگه دارید. کنترل ابزار همچنان باید در لایه ابزار یا ایجنت اعمال شود.',
              },
              {
                icon: Globe, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10', border: 'border-amber-100 dark:border-amber-500/20',
                title: 'مناسب برای متن فارسی و جریان‌های داخلی',
                desc: 'عبارت‌ها و Regex گاردریل را برای متن فارسی و الگوهای سازمانی تنظیم کنید و هزینه مصرف را با نمایش ریالی بررسی کنید.',
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

      {/* ═══ 9. Use Cases ════════════════════════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              اجرای ایجنت از گدارAI برای چه سناریوهایی مناسب است؟
            </h2>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            {[
              {
                icon: Cpu, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10',
                title: 'دستیارهای هوشمند محصول',
                desc: 'فراخوانی مدل دستیار محصول را با هویت مستقل اجرا کنید و هزینه، خطا و مسیر مقصد آن را قابل مشاهده نگه دارید.',
              },
              {
                icon: MessagesSquare, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10',
                title: 'ایجنت‌های پشتیبانی مشتری',
                desc: 'فراخوانی مدل پشتیبانی را از هویت سرویس مستقل بفرستید و گاردریل متن، هزینه و خطا را برای همان سرویس بررسی کنید.',
              },
              {
                icon: GitBranch, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10',
                title: 'اتوماسیون جریان کار',
                desc: 'فراخوانی‌های مدل در جریان چندمرحله‌ای را با شناسه درخواست به ردیابی اپلیکیشن وصل کنید و خطا یا مسیر جایگزین را ببینید.',
              },
              {
                icon: Search, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10',
                title: 'ایجنت‌های داده و پژوهش',
                desc: 'برای ایجنت‌های خلاصه‌سازی یا تحلیل، مدل، نسخه پرامپت، هزینه و خطا را کنار ارزیابی کیفیت خود نگه دارید.',
              },
              {
                icon: Building, color: 'text-rose-600 dark:text-rose-400', bg: 'bg-rose-50 dark:bg-rose-500/10',
                title: 'ایجنت‌های داخلی سازمان',
                desc: 'برای تیم‌های عملیات، مالی، فروش، پشتیبانی یا محصول، هویت سرویس و مدل‌های مجاز جدا تعریف کنید.',
              },
              {
                icon: Share2, color: 'text-teal-600 dark:text-teal-400', bg: 'bg-teal-50 dark:bg-teal-500/10',
                title: 'سامانه‌های چندایجنتی',
                desc: 'برای هر ایجنت VAT جدا بسازید تا مصرف و خطای فراخوانی‌های مدل در سطح هر هویت قابل تفکیک باشد.',
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

      {/* ═══ 10. Resources ════════════════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
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
                title: 'چطور فراخوانی مدل ایجنت را عملیاتی کنیم؟',
                desc: 'هویت سرویس، مدل مجاز، مهلت پاسخ، مسیر جایگزین، گاردریل متن و داده‌های پایش را پیش از انتشار مشخص کنید.',
              },
              {
                tag: 'مقاله', color: 'text-purple-600 dark:text-purple-400', tagBg: 'bg-purple-50 dark:bg-purple-500/10',
                title: 'چطور ردیابی ایجنت و درگاه را به هم وصل کنیم؟',
                desc: 'شناسه درخواست و رد درخواست گدارAI را کنار ردیابی چارچوب ایجنت نگه دارید تا مرز خطای مدل و اپلیکیشن روشن باشد.',
              },
              {
                tag: 'راهنما', color: 'text-emerald-600 dark:text-emerald-400', tagBg: 'bg-emerald-50 dark:bg-emerald-500/10',
                title: 'کنترل هزینه در جریان‌های ایجنتی',
                desc: 'برای هر ایجنت VAT جدا بسازید و با سقف هزینه، محدودیت نرخ، حافظه نهان و انتخاب مدل، مصرف فراخوانی‌ها را کنترل کنید.',
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

      {/* ═══ 11. FAQ ══════════════════════════════════════════════════════════ */}
      <section className="py-24 border-t border-slate-100 dark:border-slate-800/50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white" style={{ fontFamily: 'var(--font-heading)' }}>
              پرسش‌های پرتکرار
            </h2>
          </div>
          <FAQAccordion items={faqItems} />
        </div>
      </section>

      {/* ═══ 12. Final CTA ════════════════════════════════════════════════════ */}
      <section className="py-24 relative overflow-hidden">
        <div className="absolute inset-0 bg-indigo-600 dark:bg-indigo-900/40" />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            <div>
              <h2 className="text-3xl md:text-4xl font-bold text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
                فراخوانی‌های مدل ایجنت را قابل کنترل کنید
              </h2>
              <p className="text-lg text-indigo-200 mb-10 leading-relaxed">
                درگاه، مشاهده‌پذیری، گاردریل متن، مسیریابی، تلاش دوباره، مسیر جایگزین و کنترل هزینه را برای فراخوانی مدل ایجنت در یک مسیر داشته باشید.
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


