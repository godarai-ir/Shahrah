'use client';

import { motion, AnimatePresence } from 'motion/react';
import Link from 'next/link';
import { MagicCard } from '../../components/ui/MagicCard';
import { FAQAccordion } from '../../components/ui/FAQAccordion';
import {
  ArrowLeft, Zap, Shield, GitBranch, RefreshCw, Layers, BarChart3,
  Clock, DollarSign, CheckCircle, Key, Network, Activity, Database,
  Globe, CreditCard, AlertTriangle, ToggleRight, FlipHorizontal,
  FileText, SlidersHorizontal, Cpu, LayoutDashboard, Users,
  ShoppingCart, Briefcase, TrendingUp, Bot, BookOpen, ArrowRight,
  Check
} from 'lucide-react';
import { useState, useEffect } from 'react';

const fadeIn = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.5 }
};

// ─── Hero Gateway Dashboard ──────────────────────────────────────────────────
function GatewayDashboard() {
  const [requestCount, setRequestCount] = useState(12450);
  const [activeProvider, setActiveProvider] = useState(0);
  const providers = ['OpenAI', 'Anthropic', 'Google'];

  useEffect(() => {
    const t1 = setInterval(() => setRequestCount(v => v + Math.floor(Math.random() * 5) + 1), 1800);
    const t2 = setInterval(() => setActiveProvider(v => (v + 1) % providers.length), 3000);
    return () => { clearInterval(t1); clearInterval(t2); };
  }, []);

  const logs = [
    { status: '200', model: 'gpt-4o', provider: 'OpenAI', latency: '380ms', tokens: '520', cache: 'HIT', cost: '۶۲۰ ت' },
    { status: '200', model: 'claude-3.5-sonnet', provider: 'Anthropic', latency: '710ms', tokens: '1100', cache: 'MISS', cost: '۱,۳۲۰ ت' },
    { status: '429', model: 'gpt-4-turbo', provider: 'OpenAI', latency: '90ms', tokens: '—', cache: '—', cost: '۰ ت' },
    { status: '200', model: 'gemini-1.5-pro', provider: 'Google', latency: '510ms', tokens: '840', cache: 'HIT', cost: '۴۸۰ ت' },
  ];

  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl overflow-hidden">
      {/* Window bar */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-slate-200 dark:border-slate-800 bg-white/50 dark:bg-slate-900/50">
        <div className="w-3 h-3 rounded-full bg-rose-500/80" />
        <div className="w-3 h-3 rounded-full bg-amber-500/80" />
        <div className="w-3 h-3 rounded-full bg-emerald-500/80" />
        <div className="mx-auto flex items-center gap-2 px-3 py-1 rounded-md bg-slate-100 dark:bg-slate-800 text-xs text-slate-500 dark:text-slate-400 font-medium font-sans">
          <Network className="w-3 h-3" />
          GodarAI AI Gateway
        </div>
      </div>

      <div className="p-5 space-y-4">
        {/* Top stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {[
            { label: 'درخواست‌های امروز', value: requestCount.toLocaleString('fa-IR'), icon: Network, color: 'text-indigo-500', bg: 'bg-indigo-50 dark:bg-indigo-500/10' },
            { label: 'نرخ موفقیت', value: '۹۹.۹٪', icon: CheckCircle, color: 'text-emerald-500', bg: 'bg-emerald-50 dark:bg-emerald-500/10' },
            { label: 'میانگین تأخیر', value: '۲.۱s', icon: Clock, color: 'text-amber-500', bg: 'bg-amber-50 dark:bg-amber-500/10' },
            { label: 'Cache Hit Rate', value: '۳۲٪', icon: Database, color: 'text-purple-500', bg: 'bg-purple-50 dark:bg-purple-500/10' },
          ].map((stat, i) => (
            <div key={i} className="p-3 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
              <div className="flex justify-between items-start mb-2">
                <span className="text-[10px] text-slate-500 dark:text-slate-400">{stat.label}</span>
                <div className={`w-6 h-6 rounded-lg ${stat.bg} flex items-center justify-center`}>
                  <stat.icon className={`w-3 h-3 ${stat.color}`} />
                </div>
              </div>
              <div className="text-lg font-bold font-sans text-slate-900 dark:text-white">{stat.value}</div>
            </div>
          ))}
        </div>

        {/* Routing flow + Key management */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {/* Routing flow */}
          <div className="md:col-span-2 p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-xs font-bold text-slate-900 dark:text-white">Smart Routing — مسیر فعال</h3>
              <span className="flex items-center gap-1 text-[10px] text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-500/10 px-2 py-0.5 rounded-full font-medium">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse inline-block" />
                Live
              </span>
            </div>
            <div className="flex items-center justify-between gap-1 font-sans text-[10px]">
              {[
                { label: 'محصول شما', icon: LayoutDashboard, color: 'bg-slate-800 text-white dark:bg-slate-700' },
                { label: 'Gateway', icon: Network, color: 'bg-indigo-600 text-white' },
                { label: 'Routing Rules', icon: GitBranch, color: 'bg-purple-600 text-white' },
                { label: providers[activeProvider], icon: Globe, color: 'bg-emerald-600 text-white' },
              ].map((node, i, arr) => (
                <div key={i} className="flex items-center gap-1">
                  <div className={`flex flex-col items-center gap-1 min-w-0`}>
                    <div className={`w-8 h-8 rounded-xl ${node.color} flex items-center justify-center shrink-0`}>
                      <node.icon className="w-4 h-4" />
                    </div>
                    <span className="text-slate-600 dark:text-slate-400 text-center leading-tight">{node.label}</span>
                  </div>
                  {i < arr.length - 1 && (
                    <div className="flex-1 flex items-center min-w-[12px] mt-[-14px]">
                      <div className="flex-1 border-t-2 border-dashed border-indigo-200 dark:border-indigo-700" />
                      <ArrowLeft className="w-2.5 h-2.5 text-indigo-400 shrink-0" />
                    </div>
                  )}
                </div>
              ))}
            </div>

            <div className="mt-4 grid grid-cols-3 gap-2 text-[10px] font-sans">
              {[
                { label: 'Fallback', val: 'فعال', ok: true },
                { label: 'Retry', val: 'x3', ok: true },
                { label: 'Rate Limit', val: '۱۰k/min', ok: true },
              ].map((b, i) => (
                <div key={i} className="flex items-center justify-between bg-slate-50 dark:bg-slate-800/50 px-2 py-1.5 rounded-lg border border-slate-100 dark:border-slate-700/50">
                  <span className="text-slate-500 dark:text-slate-400">{b.label}</span>
                  <span className={`font-bold ${b.ok ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-500'}`}>{b.val}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Key management */}
          <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <h3 className="text-xs font-bold text-slate-900 dark:text-white mb-3 flex items-center gap-1.5">
              <Key className="w-3 h-3 text-amber-500" /> مدیریت کلیدها
            </h3>
            <div className="space-y-2 font-sans text-[10px]">
              {[
                { name: 'prod-openai-key', status: 'فعال', color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10' },
                { name: 'prod-anthropic-key', status: 'فعال', color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10' },
                { name: 'dev-gemini-key', status: 'محدود', color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10' },
                { name: 'old-openai-key', status: 'منقضی', color: 'text-rose-600 dark:text-rose-400', bg: 'bg-rose-50 dark:bg-rose-500/10' },
              ].map((key, i) => (
                <div key={i} className="flex items-center justify-between gap-2 py-1.5 border-b border-slate-50 dark:border-slate-800/50 last:border-0">
                  <div className="flex items-center gap-1.5 min-w-0">
                    <Key className="w-2.5 h-2.5 text-slate-400 shrink-0" />
                    <span className="text-slate-700 dark:text-slate-300 truncate">{key.name}</span>
                  </div>
                  <span className={`shrink-0 px-1.5 py-0.5 rounded-full font-bold ${key.bg} ${key.color}`}>{key.status}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Request log */}
        <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
          <h3 className="text-xs font-bold text-slate-900 dark:text-white mb-3">گزارش‌های رخداد درخواست (Live)</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-[10px] font-sans text-right">
              <thead>
                <tr className="text-slate-400 dark:text-slate-500 border-b border-slate-100 dark:border-slate-800">
                  {['وضعیت', 'مدل', 'تأمین‌کننده', 'تأخیر', 'توکن', 'Cache', 'هزینه'].map(h => (
                    <th key={h} className="pb-2 font-medium">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {logs.map((l, i) => (
                  <tr key={i} className="border-b border-slate-50 dark:border-slate-800/40 last:border-0">
                    <td className="py-2">
                      <span className={`px-1.5 py-0.5 rounded-full font-bold text-[9px] ${l.status === '200' ? 'bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400' : 'bg-rose-100 dark:bg-rose-500/20 text-rose-700 dark:text-rose-400'}`}>{l.status}</span>
                    </td>
                    <td className="py-2 text-slate-900 dark:text-slate-300 font-medium">{l.model}</td>
                    <td className="py-2 text-slate-500">{l.provider}</td>
                    <td className="py-2 text-slate-700 dark:text-slate-300">{l.latency}</td>
                    <td className="py-2 text-slate-500">{l.tokens}</td>
                    <td className="py-2">
                      <span className={`px-1.5 py-0.5 rounded text-[9px] font-bold ${l.cache === 'HIT' ? 'bg-purple-100 dark:bg-purple-500/20 text-purple-700 dark:text-purple-400' : l.cache === 'MISS' ? 'bg-slate-100 dark:bg-slate-800 text-slate-500' : 'text-slate-400'}`}>{l.cache}</span>
                    </td>
                    <td className="py-2 text-slate-900 dark:text-slate-300">{l.cost}</td>
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

// ─── Routing Flow Visual ─────────────────────────────────────────────────────
function RoutingFlowVisual() {
  const steps = [
    { label: 'دریافت درخواست', color: 'bg-slate-700 text-white dark:bg-slate-600', dot: 'bg-slate-400' },
    { label: 'بررسی Routing Rule', color: 'bg-indigo-600 text-white', dot: 'bg-indigo-400' },
    { label: 'انتخاب مدل اصلی', color: 'bg-purple-600 text-white', dot: 'bg-purple-400' },
    { label: 'بررسی خطا / Rate Limit', color: 'bg-amber-500 text-white', dot: 'bg-amber-400' },
    { label: 'اجرای Fallback', color: 'bg-orange-500 text-white', dot: 'bg-orange-400' },
    { label: 'ارسال پاسخ نهایی', color: 'bg-emerald-600 text-white', dot: 'bg-emerald-400' },
  ];
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs">
      <div className="flex items-center gap-1.5 mb-4">
        <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
        <span className="text-[11px] font-bold text-slate-700 dark:text-slate-300">Smart Routing Flow</span>
      </div>
      <div className="space-y-2">
        {steps.map((s, i) => (
          <div key={i} className="flex items-center gap-2">
            <div className={`w-6 h-6 rounded-lg ${s.color} flex items-center justify-center text-[10px] font-bold shrink-0`}>{i + 1}</div>
            <div className="flex-1 h-8 px-3 flex items-center rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-300 font-medium">
              {s.label}
            </div>
            {i < steps.length - 1 && <div className="w-px h-2 bg-slate-200 dark:bg-slate-700 mx-auto absolute" style={{ display: 'none' }} />}
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Cache Metrics Visual ────────────────────────────────────────────────────
function CacheMetricsVisual() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans">
      <div className="grid grid-cols-2 gap-3 mb-4">
        {[
          { label: 'Cache Hit Rate', value: '۳۲٪', color: 'text-purple-500' },
          { label: 'هزینه ذخیره‌شده', value: '۱۸٪', color: 'text-emerald-500' },
          { label: 'کاهش تأخیر', value: '~۶۰٪', color: 'text-amber-500' },
          { label: 'پاسخ‌های Cached', value: '۳,۹۸۴', color: 'text-indigo-500' },
        ].map((m, i) => (
          <div key={i} className="p-3 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <div className="text-[10px] text-slate-500 dark:text-slate-400 mb-1">{m.label}</div>
            <div className={`text-xl font-bold ${m.color}`}>{m.value}</div>
          </div>
        ))}
      </div>
      <div className="p-3 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
        <div className="text-[10px] text-slate-500 dark:text-slate-400 mb-2">نرخ Cache در طول روز</div>
        <div className="flex items-end gap-1 h-12">
          {[20, 35, 28, 45, 32, 50, 42, 58, 47, 62, 55, 68].map((h, i) => (
            <div key={i} className="flex-1 bg-purple-500/20 hover:bg-purple-500/40 rounded-t transition-all" style={{ height: `${h}%` }} />
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── Key Vault Visual ────────────────────────────────────────────────────────
function KeyVaultVisual() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans">
      <div className="flex items-center gap-2 mb-4">
        <div className="w-8 h-8 rounded-xl bg-amber-500/10 flex items-center justify-center">
          <Key className="w-4 h-4 text-amber-500" />
        </div>
        <div>
          <div className="text-xs font-bold text-slate-900 dark:text-white">API Key Vault</div>
          <div className="text-[10px] text-slate-500">مدیریت مرکزی کلیدها</div>
        </div>
      </div>
      <div className="space-y-2 text-xs">
        {[
          { name: 'openai-prod-v2', usage: '60%', status: 'فعال', statusColor: 'text-emerald-600 dark:text-emerald-400', bar: 'bg-emerald-500', w: '60%' },
          { name: 'anthropic-prod', usage: '35%', status: 'فعال', statusColor: 'text-emerald-600 dark:text-emerald-400', bar: 'bg-purple-500', w: '35%' },
          { name: 'gemini-staging', usage: '12%', status: 'محدود', statusColor: 'text-amber-600 dark:text-amber-400', bar: 'bg-amber-500', w: '12%' },
          { name: 'openai-legacy', usage: '0%', status: 'منقضی', statusColor: 'text-rose-600 dark:text-rose-400', bar: 'bg-rose-500', w: '0%' },
        ].map((k, i) => (
          <div key={i} className="p-3 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-1.5">
                <Key className="w-3 h-3 text-slate-400" />
                <span className="text-slate-700 dark:text-slate-300 font-medium">{k.name}</span>
              </div>
              <span className={`text-[10px] font-bold ${k.statusColor}`}>{k.status}</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="flex-1 h-1.5 bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden">
                <div className={`${k.bar} h-full rounded-full`} style={{ width: k.w }} />
              </div>
              <span className="text-[10px] text-slate-400">{k.usage}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Connection Diagram ──────────────────────────────────────────────────────
function ConnectionDiagram() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs">
      <div className="text-[11px] font-bold text-slate-700 dark:text-slate-300 mb-4 text-center">اتصال از یک API واحد به همه مدل‌ها</div>
      <div className="flex flex-col items-center gap-3">
        {/* App */}
        <div className="w-40 py-2 px-4 rounded-xl bg-slate-800 dark:bg-slate-700 text-white text-center font-bold text-[11px] shadow-lg">
          📱 اپلیکیشن شما
        </div>
        <div className="w-px h-4 bg-indigo-300 dark:bg-indigo-700" />
        {/* Gateway */}
        <div className="w-40 py-2 px-4 rounded-xl bg-gradient-to-l from-indigo-600 to-purple-600 text-white text-center font-bold text-[11px] shadow-lg shadow-indigo-500/20">
          🛡️ GodarAI Gateway
        </div>
        {/* Fork lines */}
        <div className="flex items-start gap-6 relative">
          <div className="absolute top-0 left-1/2 -translate-x-1/2 w-px h-3 bg-indigo-300 dark:bg-indigo-700" />
          <div className="flex gap-4 pt-3">
            {[
              { label: 'OpenAI', emoji: '🟢', bg: 'bg-emerald-50 dark:bg-emerald-500/10 border-emerald-200 dark:border-emerald-500/20' },
              { label: 'Anthropic', emoji: '🟣', bg: 'bg-purple-50 dark:bg-purple-500/10 border-purple-200 dark:border-purple-500/20' },
              { label: 'Google', emoji: '🔵', bg: 'bg-blue-50 dark:bg-blue-500/10 border-blue-200 dark:border-blue-500/20' },
            ].map((p, i) => (
              <div key={i} className={`w-20 py-2 rounded-xl ${p.bg} border text-center`}>
                <div className="text-base">{p.emoji}</div>
                <div className="text-[10px] font-bold text-slate-700 dark:text-slate-300 mt-0.5">{p.label}</div>
              </div>
            ))}
          </div>
        </div>
        {/* Modality chips */}
        <div className="flex flex-wrap gap-1.5 justify-center mt-1">
          {['Text', 'Vision', 'Audio', 'Image Gen', 'Embeddings'].map(t => (
            <span key={t} className="px-2 py-0.5 rounded-full bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400 text-[10px] font-medium">{t}</span>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── CTA Mini Dashboard ──────────────────────────────────────────────────────
function CTAMiniDashboard() {
  return (
    <div className="bg-white/10 backdrop-blur-sm border border-white/20 rounded-2xl p-5 font-sans text-sm text-white">
      <div className="flex items-center gap-2 mb-4 pb-3 border-b border-white/10">
        <Network className="w-4 h-4 text-indigo-300" />
        <span className="text-xs font-bold text-white/80">Gateway — داشبورد لحظه‌ای</span>
        <span className="mr-auto flex items-center gap-1 text-[10px] text-emerald-300 bg-emerald-500/20 px-2 py-0.5 rounded-full">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse inline-block" />
          Live
        </span>
      </div>
      <div className="grid grid-cols-2 gap-3">
        {[
          { label: '۱۲,۴۵۰ درخواست امروز', icon: Network, color: 'text-indigo-300' },
          { label: '۹۹.۹٪ نرخ موفقیت', icon: CheckCircle, color: 'text-emerald-300' },
          { label: '۲.۱ ثانیه میانگین تأخیر', icon: Clock, color: 'text-amber-300' },
          { label: '۳۲٪ Cache Hit Rate', icon: Database, color: 'text-purple-300' },
          { label: '۱۸٪ کاهش هزینه با Routing', icon: DollarSign, color: 'text-teal-300' },
          { label: 'Fallback فعال — ۳ مسیر', icon: GitBranch, color: 'text-rose-300' },
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

// ─── FAQ Data ────────────────────────────────────────────────────────────────
const faqItems = [
  {
    question: 'AI Gateway گدارAI دقیقاً چه کاری انجام می‌دهد؟',
    answer: 'گدارAI یک لایه مرکزی بین محصول شما و مدل‌های AI است. درخواست‌ها از طریق این لایه ارسال می‌شوند تا اتصال، Routing، Fallback، Retry، Cache، مدیریت کلیدها، پایش و کنترل هزینه ساده‌تر شود.',
  },
  {
    question: 'آیا گدارAI باعث افزایش تأخیر درخواست‌ها می‌شود؟',
    answer: 'هدف گدارAI این است که با Routing مناسب، Cache و مدیریت بهتر مسیرها، تأخیر را کنترل کند. در بسیاری از سناریوها، Cache و انتخاب مسیر مناسب می‌تواند پاسخ‌ها را سریع‌تر و پایدارتر کند.',
  },
  {
    question: 'اگر یک تأمین‌کننده خطا بدهد یا به Rate Limit برسد چه می‌شود؟',
    answer: 'می‌توانید قانون Fallback تعریف کنید تا درخواست به مدل یا تأمین‌کننده جایگزین منتقل شود. همچنین با Retry خودکار می‌توان خطاهای موقت را بدون دخالت کاربر مدیریت کرد.',
  },
  {
    question: 'آیا می‌توانیم مشخص کنیم کدام مدل‌ها یا تأمین‌کننده‌ها مجاز هستند؟',
    answer: 'بله. در ساختار Gateway می‌توانید دسترسی به مدل‌ها، کلیدها و مسیرهای مجاز را کنترل کنید تا تیم‌ها فقط از مسیرهای تعریف‌شده استفاده کنند.',
  },
  {
    question: 'آیا گدارAI برای اپلیکیشن‌هایی با رشد سریع مناسب است؟',
    answer: 'بله. گدارAI برای سناریوهایی طراحی می‌شود که مصرف AI به‌مرور زیاد می‌شود و تیم به کنترل متمرکز، پایش، توزیع بار و مدیریت هزینه نیاز دارد.',
  },
  {
    question: 'آیا برای استفاده از گدارAI نیاز به پرداخت ارزی دارم؟',
    answer: 'خیر. گدارAI برای تیم‌های ایرانی طراحی شده و هدف آن ساده‌کردن دسترسی به مدل‌های AI بدون نیاز به پرداخت ارزی مستقیم و با ریسک کمتر نسبت به اتصال مستقیم به سرویس‌های خارجی است.',
  },
];

// ─── Main Component ───────────────────────────────────────────────────────────
export function AIGateway() {
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
              <Network className="w-4 h-4" />
              AI Gateway برای تیم‌های ایرانی
            </motion.div>

            <motion.h1
              className="text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight mb-8 leading-tight text-slate-900 dark:text-white"
              style={{ fontFamily: 'var(--font-heading)' }}
              {...fadeIn}
            >
              یک درگاه واحد برای اتصال{' '}
              <span className="text-transparent bg-clip-text bg-gradient-to-l from-indigo-500 to-purple-500">
                امن، سریع و پایدار
              </span>{' '}
              به مدل‌های AI
            </motion.h1>

            <motion.p
              className="text-lg md:text-xl text-slate-600 dark:text-slate-400 mb-10 leading-relaxed max-w-3xl mx-auto"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
            >
              با گدارAI، درخواست‌های مدل زبانی محصول خود را از یک مسیر واحد مدیریت کنید؛ به چند مدل و ارائه‌دهنده وصل شوید، کلیدها را امن نگه دارید، هزینه‌ها را کنترل کنید و با کاهش ریسک اتصال مستقیم و پرداخت ارزی، هوش مصنوعی را وارد محصول واقعی کنید.
            </motion.p>

            <motion.div
              className="flex flex-col sm:flex-row items-center justify-center gap-4"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
            >
              <Link
                href="/pricing"
                className="w-full sm:w-auto px-8 py-4 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-lg shadow-[0_0_20px_rgba(79,70,229,0.3)] hover:shadow-[0_0_30px_rgba(79,70,229,0.5)] transition-all active:scale-95 flex items-center justify-center gap-2 group"
              >
                شروع کنید
                <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
              </Link>
              <Link
                href="/contact"
                className="w-full sm:w-auto px-8 py-4 bg-white dark:bg-slate-900 text-slate-900 dark:text-white border-2 border-slate-200 dark:border-slate-800 hover:border-indigo-600 dark:hover:border-indigo-500 rounded-xl font-bold text-lg transition-all active:scale-95 flex items-center justify-center"
              >
                درخواست دمو
              </Link>
            </motion.div>
          </div>

          {/* Hero Dashboard */}
          <motion.div
            className="mt-20 relative mx-auto w-full max-w-6xl"
            initial={{ opacity: 0, y: 40 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, delay: 0.3 }}
          >
            <div
              className="absolute inset-0 bg-gradient-to-t from-white via-transparent to-transparent dark:from-[#0B0F19] z-20 pointer-events-none h-full w-full"
              style={{ top: '65%' }}
            />
            <GatewayDashboard />
          </motion.div>
        </div>
      </section>

      {/* ═══ 2. Trust Bar ════════════════════════════════════════════════════ */}
      <section className="py-12 border-y border-slate-100 dark:border-slate-800/50 bg-slate-50/50 dark:bg-slate-900/20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-8">
            <h2 className="text-xl font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
              ساخته‌شده برای تیم‌هایی که AI را در محصول واقعی اجرا می‌کنند
            </h2>
            <p className="text-sm text-slate-600 dark:text-slate-400 max-w-2xl mx-auto">
              گدارAI به تیم‌های محصول، فنی و AI کمک می‌کند اتصال به مدل‌های زبانی را از حالت آزمایشی به زیرساخت پایدار محصول تبدیل کنند؛ با مدیریت مرکزی، کنترل هزینه، پایش لحظه‌ای و تجربه‌ای مناسب بازار ایران.
            </p>
          </div>
          <div className="flex flex-wrap justify-center gap-3">
            {[
              'استارتاپ‌های AI', 'تیم‌های Backend و Platform', 'محصولات SaaS',
              'تیم‌های Customer Support', 'سازمان‌های داده‌محور', 'تیم‌های R&D و AI Lab',
              'شرکت‌هایی که چند مدل AI را هم‌زمان تست می‌کنند',
            ].map((chip, idx) => (
              <div key={idx} className="px-4 py-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-full text-sm font-medium text-slate-700 dark:text-slate-300 shadow-sm">
                {chip}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 3. Lightning-fast AI Gateway ═══════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              AI Gateway سریع، پایدار و آماده محیط عملیاتی
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              گدارAI یک لایه مرکزی بین محصول شما و مدل‌های LLM است؛ تا درخواست‌ها را مدیریت، مسیریابی، مانیتور و بهینه کند.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Card 1 — Multi-model */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(99,102,241,0.1)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-indigo-100 dark:bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 rounded-xl flex items-center justify-center mb-6">
                  <Network className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  اتصال به چند مدل از طریق یک API واحد
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  به‌جای پیاده‌سازی جداگانه برای هر تأمین‌کننده، همه درخواست‌های AI را از یک API واحد مدیریت کنید. گدارAI کمک می‌کند مدل‌های مختلف را سریع‌تر تست، جایگزین و در محصول استفاده کنید.
                </p>
                <p className="text-xs text-indigo-600 dark:text-indigo-400 font-medium mb-6">
                  کمتر کد بزنید، سریع‌تر مدل عوض کنید، کنترل بیشتری داشته باشید.
                </p>
                <ConnectionDiagram />
              </div>
            </MagicCard>

            {/* Card 2 — Smart Routing */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(139,92,246,0.1)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-purple-100 dark:bg-purple-500/20 text-purple-600 dark:text-purple-400 rounded-xl flex items-center justify-center mb-6">
                  <GitBranch className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  پایداری بیشتر با Smart Routing
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  درخواست‌ها را بر اساس قانون‌های قابل تنظیم بین مدل‌ها و تأمین‌کننده‌ها مسیریابی کنید. اگر یک مدل کند شد، خطا داد یا به محدودیت رسید، گدارAI می‌تواند مسیر مناسب‌تری برای ادامه درخواست انتخاب کند.
                </p>
                <p className="text-xs text-purple-600 dark:text-purple-400 font-medium mb-6">
                  برای کاهش خطا، افزایش Uptime و کنترل بهتر تجربه کاربر.
                </p>
                <RoutingFlowVisual />
              </div>
            </MagicCard>

            {/* Card 3 — Caching */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(168,85,247,0.1)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-purple-100 dark:bg-purple-500/20 text-purple-600 dark:text-purple-400 rounded-xl flex items-center justify-center mb-6">
                  <Database className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  کاهش هزینه و تأخیر با Caching
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  درخواست‌های تکراری یا مشابه را با Cache ساده یا معنایی مدیریت کنید تا هزینه مصرف توکن و زمان پاسخ کاهش پیدا کند. این قابلیت برای سناریوهایی مثل تست، پاسخ‌های پرتکرار و Queryهای مشابه بسیار کاربردی است.
                </p>
                <p className="text-xs text-purple-600 dark:text-purple-400 font-medium mb-6">
                  پاسخ سریع‌تر، هزینه کمتر، فشار کمتر روی مدل‌ها.
                </p>
                <CacheMetricsVisual />
              </div>
            </MagicCard>

            {/* Card 4 — Key Management */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(245,158,11,0.1)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-amber-100 dark:bg-amber-500/20 text-amber-600 dark:text-amber-400 rounded-xl flex items-center justify-center mb-6">
                  <Key className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  مدیریت امن کلیدهای دسترسی
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  کلیدهای مدل‌ها و تأمین‌کننده‌ها را به‌جای پخش شدن در چند سرویس و چند مخزن کد، در یک محیط مرکزی مدیریت کنید. دسترسی‌ها را کنترل، کلیدها را غیرفعال یا جایگزین کنید و مصرف هر کلید را مانیتور کنید.
                </p>
                <p className="text-xs text-amber-600 dark:text-amber-400 font-medium mb-6">
                  امنیت بیشتر برای تیم‌هایی که چند توسعه‌دهنده، چند محیط و چند محصول دارند.
                </p>
                <KeyVaultVisual />
              </div>
            </MagicCard>
          </div>
        </div>
      </section>

      {/* ═══ 4. Advanced Capabilities ════════════════════════════════════════ */}
      <section className="py-24 bg-slate-900 dark:bg-[#060913] text-white border-y border-slate-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              قابلیت‌های پیشرفته، آماده استفاده
            </h2>
            <p className="text-lg text-slate-400">
              گدارAI فقط یک اتصال ساده به مدل نیست؛ یک لایه عملیاتی برای اجرای پایدار AI در محصول است.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {/* Batching */}
            <div className="bg-slate-800/50 border border-slate-700/50 rounded-2xl p-8">
              <div className="w-12 h-12 bg-indigo-500/20 text-indigo-400 rounded-xl flex items-center justify-center mb-6">
                <Layers className="w-6 h-6" />
              </div>
              <h3 className="text-lg font-bold text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                Batching هوشمند برای حجم بالای درخواست‌ها
              </h3>
              <p className="text-slate-400 text-sm leading-relaxed mb-6">
                درخواست‌های زیاد را با روش‌های Batch مدیریت کنید تا فشار روی زیرساخت و هزینه پردازش کاهش پیدا کند؛ بدون اینکه عملکرد Real-time محصول آسیب ببیند.
              </p>
              {/* Mini batch animation */}
              <div className="bg-slate-900/50 border border-slate-700/50 rounded-xl p-4 font-sans text-xs">
                <div className="flex items-center justify-between mb-3 text-slate-400">
                  <span>درخواست‌های پراکنده</span>
                  <span>Batch واحد</span>
                </div>
                <div className="flex items-center gap-3">
                  <div className="flex flex-col gap-1">
                    {[1, 2, 3, 4].map(i => (
                      <div key={i} className="w-6 h-2 rounded bg-indigo-500/40" />
                    ))}
                  </div>
                  <ArrowLeft className="w-4 h-4 text-slate-500" />
                  <div className="flex-1 h-10 rounded-lg bg-indigo-600/30 border border-indigo-500/30 flex items-center justify-center text-indigo-400 font-bold">
                    Batch
                  </div>
                  <ArrowLeft className="w-4 h-4 text-slate-500" />
                  <div className="flex flex-col gap-1">
                    {[1, 2, 3, 4].map(i => (
                      <div key={i} className="w-6 h-2 rounded bg-emerald-500/40" />
                    ))}
                  </div>
                </div>
              </div>
            </div>

            {/* Fine-tuning */}
            <div className="bg-slate-800/50 border border-slate-700/50 rounded-2xl p-8">
              <div className="w-12 h-12 bg-purple-500/20 text-purple-400 rounded-xl flex items-center justify-center mb-6">
                <SlidersHorizontal className="w-6 h-6" />
              </div>
              <h3 className="text-lg font-bold text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                Fine-tuning از مسیر یکپارچه
              </h3>
              <p className="text-slate-400 text-sm leading-relaxed mb-6">
                اگر تیم شما از Fine-tuning استفاده می‌کند، گدارAI می‌تواند مسیر مدیریت مدل‌های سفارشی و مدل‌های پایه را ساده‌تر کند تا همه چیز از یک لایه مرکزی قابل کنترل باشد.
              </p>
              <div className="bg-slate-900/50 border border-slate-700/50 rounded-xl p-4 font-sans text-xs space-y-2">
                {[
                  { label: 'مدل پایه (gpt-4o)', badge: 'Base', color: 'text-slate-400 bg-slate-700' },
                  { label: 'مدل Fine-tuned', badge: 'Custom', color: 'text-purple-400 bg-purple-500/20' },
                ].map((row, i) => (
                  <div key={i} className="flex items-center justify-between bg-slate-800 rounded-lg px-3 py-2">
                    <span className="text-slate-300">{row.label}</span>
                    <span className={`px-2 py-0.5 rounded-full font-bold ${row.color}`}>{row.badge}</span>
                  </div>
                ))}
                <div className="flex gap-2 mt-3">
                  {[['کیفیت', '↑۲۸٪', 'text-emerald-400'], ['هزینه', '↓۱۵٪', 'text-emerald-400'], ['تأخیر', '↓۱۰٪', 'text-emerald-400']].map(([l, v, c], i) => (
                    <div key={i} className="flex-1 bg-slate-800 rounded-lg p-2 text-center">
                      <div className="text-slate-400 text-[10px]">{l}</div>
                      <div className={`font-bold ${c}`}>{v}</div>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            {/* Multimodal */}
            <div className="bg-slate-800/50 border border-slate-700/50 rounded-2xl p-8">
              <div className="w-12 h-12 bg-emerald-500/20 text-emerald-400 rounded-xl flex items-center justify-center mb-6">
                <Cpu className="w-6 h-6" />
              </div>
              <h3 className="text-lg font-bold text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                مدیریت مدل‌های چندحالته
              </h3>
              <p className="text-slate-400 text-sm leading-relaxed mb-6">
                برای سناریوهای متنی، تصویری، صوتی یا Vision، گدارAI می‌تواند به‌عنوان مسیر مرکزی ارتباط با مدل‌ها و تأمین‌کننده‌های مختلف عمل کند.
              </p>
              <div className="flex flex-wrap gap-2">
                {[
                  { label: 'Text', icon: FileText, color: 'bg-indigo-500/20 text-indigo-400 border-indigo-500/30' },
                  { label: 'Vision', icon: Activity, color: 'bg-purple-500/20 text-purple-400 border-purple-500/30' },
                  { label: 'Audio', icon: Activity, color: 'bg-pink-500/20 text-pink-400 border-pink-500/30' },
                  { label: 'Image Generation', icon: Activity, color: 'bg-amber-500/20 text-amber-400 border-amber-500/30' },
                  { label: 'Embeddings', icon: Database, color: 'bg-teal-500/20 text-teal-400 border-teal-500/30' },
                ].map((chip, i) => (
                  <span key={i} className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full border text-xs font-bold ${chip.color}`}>
                    {chip.label}
                  </span>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ═══ 5. Production AI Impact ═════════════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              زیرساختی برای AI در مقیاس محیط عملیاتی
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              وقتی AI وارد محصول واقعی می‌شود، فقط پاسخ گرفتن از مدل کافی نیست. باید پایداری، هزینه، سرعت و امنیت را هم‌زمان مدیریت کنید.
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {[
              {
                icon: Network, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10',
                value: 'میلیون‌ها', sub: 'درخواست در ماه',
                label: 'درخواست‌های پردازش‌شده',
                desc: 'گدارAI باید طوری طراحی شود که بتواند رشد سریع مصرف AI را در محصول‌های واقعی پشتیبانی کند.',
              },
              {
                icon: CheckCircle, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10',
                value: '۹۹.۹٪', sub: 'Uptime هدف',
                label: 'Uptime بالا',
                desc: 'با Routing، Fallback و Retry، تجربه کاربر حتی هنگام خطای مدل یا تأمین‌کننده پایدارتر می‌ماند.',
              },
              {
                icon: DollarSign, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10',
                value: '↓۱۸٪', sub: 'کاهش هزینه با Routing',
                label: 'کاهش هزینه',
                desc: 'با Caching، انتخاب مدل مناسب و کنترل مصرف، هزینه‌های AI قابل پیش‌بینی‌تر و قابل مدیریت‌تر می‌شود.',
              },
              {
                icon: Zap, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10',
                value: '↓۶۰٪', sub: 'کاهش تأخیر با Cache',
                label: 'Latency کمتر',
                desc: 'با انتخاب مسیر مناسب و کاهش درخواست‌های غیرضروری، زمان پاسخ برای کاربران نهایی بهتر می‌شود.',
              },
            ].map((card, i) => (
              <motion.div
                key={i}
                className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 flex flex-col"
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: i * 0.1 }}
              >
                <div className={`w-12 h-12 ${card.bg} ${card.color} rounded-xl flex items-center justify-center mb-4`}>
                  <card.icon className="w-6 h-6" />
                </div>
                <div className={`text-3xl font-bold font-sans ${card.color} mb-1`}>{card.value}</div>
                <div className="text-xs text-slate-500 dark:text-slate-400 font-medium mb-3 font-sans">{card.sub}</div>
                <h3 className="font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>{card.label}</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed flex-1">{card.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 6. Pilot to Production ══════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              از Prototype تا Production، سریع‌تر و امن‌تر
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              گدارAI به تیم‌ها کمک می‌کند AI app خود را با قابلیت‌های عملیاتی مثل Routing، Fallback، Retry، Load Balancing و مدیریت فایل‌ها پایدارتر کنند.
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            {[
              {
                icon: GitBranch, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10',
                title: 'Routing شرطی',
                desc: 'درخواست‌ها را بر اساس شرط‌های مشخص مثل نوع کاربر، هزینه، مدل، تأمین‌کننده، Latency یا نوع فیچر به مسیر مناسب ارسال کنید.',
              },
              {
                icon: Cpu, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10',
                title: 'طراحی‌شده برای مدل‌های چندحالته',
                desc: 'در کنار مدل‌های متنی، از سناریوهای Vision، صوت، تصویر و سایر مدل‌های AI نیز در یک مسیر واحد پشتیبانی کنید.',
              },
              {
                icon: FlipHorizontal, color: 'text-rose-600 dark:text-rose-400', bg: 'bg-rose-50 dark:bg-rose-500/10',
                title: 'Fallback هنگام خطا',
                desc: 'اگر مدل اصلی خطا داد، کند شد یا به محدودیت رسید، درخواست را به مدل یا تأمین‌کننده جایگزین منتقل کنید تا تجربه کاربر قطع نشود.',
              },
              {
                icon: RefreshCw, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10',
                title: 'Retry خودکار',
                desc: 'درخواست‌های ناموفق را با قانون‌های کنترل‌شده دوباره ارسال کنید تا خطاهای موقت باعث شکست تجربه کاربر نشوند.',
              },
              {
                icon: BarChart3, color: 'text-teal-600 dark:text-teal-400', bg: 'bg-teal-50 dark:bg-teal-500/10',
                title: 'توزیع بار بین مدل‌ها',
                desc: 'درخواست‌ها را بین مدل‌ها یا تأمین‌کننده‌های مختلف توزیع کنید تا فشار روی یک مسیر خاص کاهش پیدا کند و عملکرد پایدارتر شود.',
              },
              {
                icon: Activity, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10',
                title: 'پایش APIهای Real-time',
                desc: 'درخواست‌های Real-time را همراه با هزینه، وضعیت، خطاها، تأخیر و تخطی از Guardrailها ثبت و بررسی کنید.',
              },
              {
                icon: ToggleRight, color: 'text-yellow-600 dark:text-yellow-400', bg: 'bg-yellow-50 dark:bg-yellow-500/10',
                title: 'Canary Testing برای مدل‌ها و Promptها',
                desc: 'مدل‌ها، Promptها یا مسیرهای جدید را روی بخشی از ترافیک تست کنید، بدون اینکه کل کاربران تحت تأثیر قرار بگیرند.',
              },
              {
                icon: Clock, color: 'text-orange-600 dark:text-orange-400', bg: 'bg-orange-50 dark:bg-orange-500/10',
                title: 'مدیریت Timeout درخواست‌ها',
                desc: 'برای درخواست‌هایی که بیش از حد طول می‌کشند، قانون توقف تعریف کنید و در صورت نیاز مسیر جایگزین یا درخواست جدید اجرا کنید.',
              },
              {
                icon: FileText, color: 'text-sky-600 dark:text-sky-400', bg: 'bg-sky-50 dark:bg-sky-500/10',
                title: 'پشتیبانی از فایل‌ها',
                desc: 'فایل‌ها را به Gateway متصل کنید و از محتوای آن‌ها در درخواست‌ها و جریان‌های AI استفاده کنید.',
              },
            ].map((feat, i) => (
              <motion.div
                key={i}
                className="bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 hover:border-indigo-200 dark:hover:border-indigo-800 transition-colors group"
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: (i % 3) * 0.1 }}
              >
                <div className={`w-10 h-10 ${feat.bg} ${feat.color} rounded-xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform`}>
                  <feat.icon className="w-5 h-5" />
                </div>
                <h3 className="font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                  {feat.title}
                </h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">{feat.desc}</p>
              </motion.div>
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
              چرا AI Gateway گدارAI برای تیم‌های ایرانی مهم‌تر است؟
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              برای تیم‌های ایرانی، اتصال مستقیم به مدل‌های خارجی فقط یک مسئله فنی نیست؛ مسئله دسترسی، پرداخت، پایداری، هزینه و ریسک قطع سرویس هم هست.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {[
              {
                icon: Shield, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10', border: 'border-indigo-100 dark:border-indigo-500/20',
                title: 'کاهش ریسک محدودیت‌های اتصال مستقیم',
                desc: 'با گدارAI، تیم شما برای استفاده از مدل‌های AI خارجی کمتر درگیر محدودیت‌های تحریم، مسدود شدن حساب یا قطع دسترسی مستقیم می‌شود.',
              },
              {
                icon: CreditCard, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10', border: 'border-emerald-100 dark:border-emerald-500/20',
                title: 'بدون پرداخت ارزی مستقیم',
                desc: 'به‌جای درگیری با کارت اعتباری خارجی، پرداخت دلاری و حساب‌های بین‌المللی، مصرف AI را از طریق گدارAI و متناسب با نیاز بازار ایران مدیریت کنید.',
              },
              {
                icon: Network, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10', border: 'border-purple-100 dark:border-purple-500/20',
                title: 'یک لایه مرکزی برای چند مدل و تأمین‌کننده',
                desc: 'به‌جای اتصال مستقیم و پراکنده به چند سرویس، گدارAI یک مسیر واحد برای مدیریت مدل‌ها، کلیدها، هزینه، Routing و Observability فراهم می‌کند.',
              },
              {
                icon: Globe, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10', border: 'border-amber-100 dark:border-amber-500/20',
                title: 'آماده برای تیم‌های فارسی‌زبان',
                desc: 'داشبوردها، گزارش‌ها و تجربه استفاده در گدارAI برای تیم‌هایی طراحی می‌شود که محصول فارسی، کاربر ایرانی و نیازهای عملیاتی داخل ایران دارند.',
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
              AI Gateway گدارAI برای چه سناریوهایی مناسب است؟
            </h2>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            {[
              {
                icon: ShoppingCart, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10',
                title: 'محصولات SaaS با قابلیت AI',
                desc: 'اگر در محصول خود قابلیت‌هایی مثل چت‌بات، تولید محتوا، تحلیل متن یا دستیار هوشمند دارید، گدارAI کمک می‌کند اتصال به مدل‌ها پایدارتر و قابل‌کنترل‌تر شود.',
              },
              {
                icon: Layers, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10',
                title: 'تیم‌های Backend و Platform',
                desc: 'زیرساخت AI را از چند اتصال پراکنده به یک Gateway مرکزی تبدیل کنید و مدیریت Routing، Retry، Fallback، کلیدها و پایش را ساده‌تر کنید.',
              },
              {
                icon: TrendingUp, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10',
                title: 'تیم‌های محصول و رشد',
                desc: 'مدل‌ها و Promptهای مختلف را سریع‌تر تست کنید، Canary Testing انجام دهید و بفهمید کدام مسیر تجربه بهتری برای کاربران ایجاد می‌کند.',
              },
              {
                icon: Briefcase, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10',
                title: 'تیم‌های مالی و مدیریت',
                desc: 'مصرف AI را قابل مشاهده و قابل کنترل کنید، هزینه را در سطح مدل، تیم، کاربر یا فیچر ببینید و تصمیم‌های اقتصادی‌تری بگیرید.',
              },
              {
                icon: Bot, color: 'text-rose-600 dark:text-rose-400', bg: 'bg-rose-50 dark:bg-rose-500/10',
                title: 'AI Agentها و Workflowهای چندمرحله‌ای',
                desc: 'برای Agentها، Tool Callها و Workflowهای پیچیده، یک مسیر عملیاتی پایدار داشته باشید که خطاها، Timeoutها و جایگزینی مدل‌ها را مدیریت کند.',
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

      {/* 10. FAQ Section */}
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
                AI stack محصول خود را از یک Gateway واحد مدیریت کنید
              </h2>
              <p className="text-lg text-indigo-200 mb-10 leading-relaxed">
                با گدارAI، اتصال به مدل‌ها، مدیریت کلیدها، Routing، Fallback، Retry، Cache و پایش را در یک لایه مرکزی داشته باشید؛ با کاهش ریسک محدودیت‌های اتصال مستقیم و بدون نیاز به پرداخت ارزی مستقیم.
              </p>
              <div className="flex flex-col sm:flex-row items-start gap-4">
                <Link
                  href="/pricing"
                  className="w-full sm:w-auto px-8 py-4 bg-white text-indigo-700 hover:bg-indigo-50 rounded-xl font-bold text-lg shadow-[0_0_30px_rgba(0,0,0,0.2)] transition-all active:scale-95 flex items-center justify-center gap-2 group"
                >
                  شروع کنید
                  <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
                </Link>
                <Link
                  href="/contact"
                  className="w-full sm:w-auto px-8 py-4 bg-white/10 hover:bg-white/20 text-white border border-white/20 hover:border-white/40 rounded-xl font-bold text-lg transition-all active:scale-95 flex items-center justify-center"
                >
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


