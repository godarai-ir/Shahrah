'use client';

import { motion } from 'motion/react';
import Link from 'next/link';
import { MagicCard } from '../../components/ui/MagicCard';
import { FAQAccordion } from '../../components/ui/FAQAccordion';
import {
  Shield, Key, Lock, Eye, EyeOff, AlertTriangle, CheckCircle,
  Users, FileText, Globe, CreditCard, ArrowLeft, Database,
  ShieldCheck, ShieldAlert, Layers, ToggleRight, Activity,
  Briefcase, Bot, BookOpen, Network, Check, Settings,
  ClipboardList, Filter, GitBranch, TrendingUp, ShoppingCart,
  Clock, Terminal
} from 'lucide-react';
import { useState, useEffect } from 'react';

const fadeIn = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.5 },
};

// ─── Hero Security Dashboard ─────────────────────────────────────────────────
function SecurityDashboard() {
  const [blockedCount, setBlockedCount] = useState(32);
  const [redactedCount, setRedactedCount] = useState(1240);

  useEffect(() => {
    const t1 = setInterval(() => setBlockedCount(v => v + (Math.random() > 0.7 ? 1 : 0)), 3000);
    const t2 = setInterval(() => setRedactedCount(v => v + Math.floor(Math.random() * 3)), 2000);
    return () => { clearInterval(t1); clearInterval(t2); };
  }, []);

  const virtualKeys = [
    { name: 'prod-team-alpha', env: 'Production', models: 'gpt-4o', status: 'فعال', statusColor: 'text-emerald-600 dark:text-emerald-400', statusBg: 'bg-emerald-50 dark:bg-emerald-500/10', usage: '68%', bar: 'bg-emerald-500' },
    { name: 'dev-backend-v2', env: 'Development', models: 'gpt-3.5-turbo', status: 'فعال', statusColor: 'text-emerald-600 dark:text-emerald-400', statusBg: 'bg-emerald-50 dark:bg-emerald-500/10', usage: '22%', bar: 'bg-indigo-500' },
    { name: 'staging-support', env: 'Staging', models: 'claude-3', status: 'محدود', statusColor: 'text-amber-600 dark:text-amber-400', statusBg: 'bg-amber-50 dark:bg-amber-500/10', usage: '45%', bar: 'bg-amber-500' },
    { name: 'legacy-openai-key', env: 'Production', models: 'gpt-4', status: 'منقضی', statusColor: 'text-rose-600 dark:text-rose-400', statusBg: 'bg-rose-50 dark:bg-rose-500/10', usage: '0%', bar: 'bg-rose-500' },
  ];

  const auditLogs = [
    { time: '14:32:01', user: 'ali@co.ir', action: 'Key Created', key: 'prod-alpha', status: 'موفق', ok: true },
    { time: '14:31:45', user: 'system', action: 'Request Blocked', key: 'dev-v2', status: 'مسدود', ok: false },
    { time: '14:31:12', user: 'sara@co.ir', action: 'Policy Updated', key: 'staging', status: 'موفق', ok: true },
    { time: '14:30:58', user: 'bot-agent', action: 'Data Redacted', key: 'prod-alpha', status: 'Redacted', ok: true },
  ];

  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl overflow-hidden">
      {/* Window bar */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-slate-200 dark:border-slate-800 bg-white/50 dark:bg-slate-900/50">
        <div className="w-3 h-3 rounded-full bg-rose-500/80" />
        <div className="w-3 h-3 rounded-full bg-amber-500/80" />
        <div className="w-3 h-3 rounded-full bg-emerald-500/80" />
        <div className="mx-auto flex items-center gap-2 px-3 py-1 rounded-md bg-slate-100 dark:bg-slate-800 text-xs text-slate-500 dark:text-slate-400 font-medium font-sans">
          <Shield className="w-3 h-3" />
          GodarAI Security & Compliance
        </div>
      </div>

      <div className="p-5 space-y-4">
        {/* Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {[
            { label: 'Virtual Keys فعال', value: '۲۴', icon: Key, color: 'text-indigo-500', bg: 'bg-indigo-50 dark:bg-indigo-500/10' },
            { label: 'Policy Coverage', value: '۹۸٪', icon: ShieldCheck, color: 'text-emerald-500', bg: 'bg-emerald-50 dark:bg-emerald-500/10' },
            { label: 'درخواست‌های Redacted', value: redactedCount.toLocaleString('fa-IR'), icon: EyeOff, color: 'text-purple-500', bg: 'bg-purple-50 dark:bg-purple-500/10' },
            { label: 'درخواست‌های مسدودشده', value: blockedCount.toLocaleString('fa-IR'), icon: ShieldAlert, color: 'text-rose-500', bg: 'bg-rose-50 dark:bg-rose-500/10' },
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

        {/* Virtual Keys + Access Control */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {/* Virtual Keys */}
          <div className="md:col-span-2 p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <div className="flex justify-between items-center mb-3">
              <h3 className="text-xs font-bold text-slate-900 dark:text-white flex items-center gap-1.5">
                <Key className="w-3 h-3 text-indigo-500" /> Virtual Keys
              </h3>
              <span className="text-[10px] font-bold text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-500/10 px-2 py-0.5 rounded-full">
                ۲۴ کلید
              </span>
            </div>
            <div className="space-y-2 font-sans text-[10px]">
              {virtualKeys.map((k, i) => (
                <div key={i} className="flex items-center gap-2 bg-slate-50 dark:bg-slate-800/40 rounded-lg px-3 py-2">
                  <Key className="w-3 h-3 text-slate-400 shrink-0" />
                  <span className="text-slate-700 dark:text-slate-300 font-medium w-32 truncate">{k.name}</span>
                  <span className="text-slate-400 w-20 truncate">{k.env}</span>
                  <div className="flex-1 h-1.5 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden">
                    <div className={`${k.bar} h-full rounded-full`} style={{ width: k.usage }} />
                  </div>
                  <span className="text-slate-400 w-8">{k.usage}</span>
                  <span className={`shrink-0 px-1.5 py-0.5 rounded-full font-bold ${k.statusBg} ${k.statusColor}`}>{k.status}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Inbound Rules */}
          <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <h3 className="text-xs font-bold text-slate-900 dark:text-white mb-3 flex items-center gap-1.5">
              <Filter className="w-3 h-3 text-purple-500" /> Inbound Rules
            </h3>
            <div className="space-y-2 font-sans text-[10px]">
              {[
                { rule: 'IP Allowlist فعال', ok: true },
                { rule: 'Geo Block: خارج از ایران', ok: true },
                { rule: 'Dev → Prod مسدود', ok: true },
                { rule: 'درخواست بدون Metadata', ok: false },
                { rule: 'هشدار ترافیک مشکوک', ok: true },
              ].map((r, i) => (
                <div key={i} className="flex items-center gap-2 py-1.5 border-b border-slate-50 dark:border-slate-800/50 last:border-0">
                  <div className={`w-4 h-4 rounded-full flex items-center justify-center shrink-0 ${r.ok ? 'bg-emerald-100 dark:bg-emerald-500/20' : 'bg-rose-100 dark:bg-rose-500/20'}`}>
                    {r.ok
                      ? <Check className="w-2.5 h-2.5 text-emerald-600 dark:text-emerald-400" />
                      : <AlertTriangle className="w-2.5 h-2.5 text-rose-600 dark:text-rose-400" />
                    }
                  </div>
                  <span className="text-slate-700 dark:text-slate-300 leading-tight">{r.rule}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Audit Log */}
        <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
          <div className="flex justify-between items-center mb-3">
            <h3 className="text-xs font-bold text-slate-900 dark:text-white flex items-center gap-1.5">
              <ClipboardList className="w-3 h-3 text-emerald-500" /> Audit Logs (Live)
            </h3>
            <span className="flex items-center gap-1 text-[10px] text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-500/10 px-2 py-0.5 rounded-full font-medium">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse inline-block" />
              Live
            </span>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-[10px] font-sans text-right">
              <thead>
                <tr className="text-slate-400 dark:text-slate-500 border-b border-slate-100 dark:border-slate-800">
                  {['زمان', 'کاربر', 'اکشن', 'کلید', 'وضعیت'].map(h => (
                    <th key={h} className="pb-2 font-medium">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {auditLogs.map((l, i) => (
                  <tr key={i} className="border-b border-slate-50 dark:border-slate-800/40 last:border-0">
                    <td className="py-2 text-slate-500 font-sans">{l.time}</td>
                    <td className="py-2 text-slate-700 dark:text-slate-300 font-medium">{l.user}</td>
                    <td className="py-2 text-slate-600 dark:text-slate-400 font-sans">{l.action}</td>
                    <td className="py-2 text-slate-500 font-sans">{l.key}</td>
                    <td className="py-2">
                      <span className={`px-1.5 py-0.5 rounded-full font-bold text-[9px] ${
                        l.ok
                          ? 'bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400'
                          : 'bg-rose-100 dark:bg-rose-500/20 text-rose-700 dark:text-rose-400'
                      }`}>{l.status}</span>
                    </td>
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

// ─── Virtual Keys Panel ──────────────────────────────────────────────────────
function VirtualKeysPanel() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans">
      <div className="flex items-center gap-2 mb-4">
        <div className="w-7 h-7 rounded-lg bg-indigo-100 dark:bg-indigo-500/20 flex items-center justify-center">
          <Key className="w-3.5 h-3.5 text-indigo-600 dark:text-indigo-400" />
        </div>
        <span className="text-xs font-bold text-slate-800 dark:text-white">Virtual Key Manager</span>
      </div>
      <div className="space-y-2 text-xs">
        {[
          { name: 'prod-main-v3', env: 'Production', models: 'gpt-4o', limit: '۱۰k/روز', status: 'فعال', sc: 'text-emerald-600 dark:text-emerald-400', sb: 'bg-emerald-50 dark:bg-emerald-500/10' },
          { name: 'dev-team-alpha', env: 'Development', models: 'gpt-3.5', limit: '۲k/روز', status: 'فعال', sc: 'text-emerald-600 dark:text-emerald-400', sb: 'bg-emerald-50 dark:bg-emerald-500/10' },
          { name: 'staging-test', env: 'Staging', models: 'claude-3', limit: '۵k/روز', status: 'محدود', sc: 'text-amber-600 dark:text-amber-400', sb: 'bg-amber-50 dark:bg-amber-500/10' },
          { name: 'old-integration', env: 'Production', models: 'gpt-4', limit: '—', status: 'منقضی', sc: 'text-rose-600 dark:text-rose-400', sb: 'bg-rose-50 dark:bg-rose-500/10' },
        ].map((k, i) => (
          <div key={i} className="p-3 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-1.5">
                <Key className="w-3 h-3 text-slate-400" />
                <span className="font-medium text-slate-800 dark:text-white">{k.name}</span>
              </div>
              <span className={`px-2 py-0.5 rounded-full font-bold text-[10px] ${k.sb} ${k.sc}`}>{k.status}</span>
            </div>
            <div className="flex gap-3 text-[10px] text-slate-500">
              <span>محیط: <span className="text-slate-700 dark:text-slate-300 font-medium">{k.env}</span></span>
              <span>مدل: <span className="font-sans text-slate-700 dark:text-slate-300 font-medium">{k.models}</span></span>
              <span>سقف: <span className="text-slate-700 dark:text-slate-300 font-medium">{k.limit}</span></span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Inbound Rules Builder ───────────────────────────────────────────────────
function InboundRulesBuilder() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs">
      <div className="flex items-center gap-2 mb-4">
        <div className="w-7 h-7 rounded-lg bg-purple-100 dark:bg-purple-500/20 flex items-center justify-center">
          <Filter className="w-3.5 h-3.5 text-purple-600 dark:text-purple-400" />
        </div>
        <span className="text-xs font-bold text-slate-800 dark:text-white">Inbound Rules</span>
      </div>
      <div className="space-y-2">
        {[
          { num: 1, rule: 'فقط IPهای مشخص‌شده اجازه اتصال دارند', type: 'ALLOW', ok: true },
          { num: 2, rule: 'درخواست‌های خارج از کشور مجاز نیستند', type: 'BLOCK', ok: false },
          { num: 3, rule: 'محیط Dev به مدل‌های Production دسترسی ندارد', type: 'DENY', ok: false },
          { num: 4, rule: 'درخواست‌های بدون Metadata مسدود شوند', type: 'BLOCK', ok: false },
          { num: 5, rule: 'درخواست‌های مشکوک Flag شوند', type: 'FLAG', ok: true },
        ].map((r) => (
          <div key={r.num} className="flex items-center gap-2 p-2.5 rounded-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
            <span className="w-5 h-5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-500 flex items-center justify-center text-[10px] font-bold shrink-0">{r.num}</span>
            <span className="flex-1 text-slate-700 dark:text-slate-300 leading-tight">{r.rule}</span>
            <span className={`shrink-0 px-2 py-0.5 rounded font-bold text-[9px] font-sans ${
              r.type === 'ALLOW' ? 'bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400' :
              r.type === 'FLAG'  ? 'bg-amber-100 dark:bg-amber-500/20 text-amber-700 dark:text-amber-400' :
              'bg-rose-100 dark:bg-rose-500/20 text-rose-700 dark:text-rose-400'
            }`}>{r.type}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Data Redaction Visual ───────────────────────────────────────────────────
function DataRedactionVisual() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs space-y-4">
      <div className="flex items-center gap-2 mb-1">
        <div className="w-7 h-7 rounded-lg bg-rose-100 dark:bg-rose-500/20 flex items-center justify-center">
          <EyeOff className="w-3.5 h-3.5 text-rose-600 dark:text-rose-400" />
        </div>
        <span className="text-xs font-bold text-slate-800 dark:text-white">Data Redaction</span>
      </div>
      {/* Before */}
      <div className="p-3 rounded-xl bg-rose-50 dark:bg-rose-500/10 border border-rose-200 dark:border-rose-500/20">
        <div className="text-[10px] text-rose-600 dark:text-rose-400 font-bold mb-1.5">قبل از Redaction</div>
        <p className="text-slate-700 dark:text-slate-300 leading-relaxed">
          شماره تماس کاربر <span className="bg-rose-200 dark:bg-rose-500/30 text-rose-800 dark:text-rose-200 px-1 rounded font-medium">0912XXXXXXX</span> و ایمیل <span className="bg-rose-200 dark:bg-rose-500/30 text-rose-800 dark:text-rose-200 px-1 rounded font-medium">user@example.com</span> است.
        </p>
      </div>
      {/* After */}
      <div className="p-3 rounded-xl bg-emerald-50 dark:bg-emerald-500/10 border border-emerald-200 dark:border-emerald-500/20">
        <div className="text-[10px] text-emerald-600 dark:text-emerald-400 font-bold mb-1.5">بعد از Redaction</div>
        <p className="text-slate-700 dark:text-slate-300 leading-relaxed">
          شماره تماس کاربر <span className="bg-slate-200 dark:bg-slate-700 text-slate-500 dark:text-slate-400 px-1 rounded font-sans font-medium">[PHONE_REDACTED]</span> و ایمیل <span className="bg-slate-200 dark:bg-slate-700 text-slate-500 dark:text-slate-400 px-1 rounded font-sans font-medium">[EMAIL_REDACTED]</span> است.
        </p>
      </div>
      {/* Chips */}
      <div className="flex flex-wrap gap-1.5">
        {['Phone', 'Email', 'National ID', 'API Key', 'Token', 'Personal Data', 'Custom Pattern'].map(c => (
          <span key={c} className="px-2 py-0.5 rounded-full bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 font-sans font-medium text-[10px]">{c}</span>
        ))}
      </div>
    </div>
  );
}

// ─── Compliance Status Card ──────────────────────────────────────────────────
function ComplianceStatusCard() {
  return (
    <div className="bg-slate-50 dark:bg-[#121827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 font-sans text-xs">
      <div className="flex items-center gap-2 mb-4">
        <div className="w-7 h-7 rounded-lg bg-emerald-100 dark:bg-emerald-500/20 flex items-center justify-center">
          <ShieldCheck className="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
        </div>
        <span className="text-xs font-bold text-slate-800 dark:text-white">Compliance Status</span>
        <span className="mr-auto text-[10px] font-bold text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-500/10 px-2 py-0.5 rounded-full">۹۸٪</span>
      </div>
      <div className="space-y-2">
        {[
          { label: 'Audit Logs', status: 'فعال', ok: true },
          { label: 'Access Control', status: 'فعال', ok: true },
          { label: 'Data Redaction', status: 'فعال', ok: true },
          { label: 'Key Rotation', status: 'فعال', ok: true },
          { label: 'Policy Coverage', status: '۹۸٪', ok: true },
          { label: 'Risk Alerts', status: 'فعال', ok: true },
        ].map((item, i) => (
          <div key={i} className="flex items-center justify-between py-1.5 border-b border-slate-100 dark:border-slate-800/50 last:border-0">
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded-full bg-emerald-100 dark:bg-emerald-500/20 flex items-center justify-center">
                <Check className="w-2.5 h-2.5 text-emerald-600 dark:text-emerald-400" />
              </div>
              <span className="text-slate-700 dark:text-slate-300">{item.label}</span>
            </div>
            <span className="font-bold text-emerald-600 dark:text-emerald-400">{item.status}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Audit Log Table Visual ──────────────────────────────────────────────────
function AuditLogTable() {
  return (
    <div className="bg-slate-900 dark:bg-slate-950 border border-slate-800 rounded-2xl p-5 font-sans text-xs">
      <div className="flex items-center justify-between mb-4">
        <span className="text-sm font-bold text-white flex items-center gap-2">
          <ClipboardList className="w-4 h-4 text-emerald-400" /> Audit Logs
        </span>
        <span className="flex items-center gap-1 text-[10px] text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded-full">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse inline-block" />
          Live
        </span>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-right">
          <thead>
            <tr className="text-slate-500 border-b border-slate-800">
              {['زمان', 'کاربر', 'اکشن', 'کلید', 'مدل', 'وضعیت', 'IP'].map(h => (
                <th key={h} className="pb-2 font-medium text-[10px]">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {[
              { time: '14:35:02', user: 'ali@co', action: 'Key Created', key: 'prod-v3', model: 'gpt-4o', status: 'موفق', ok: true, ip: '185.x.x.x' },
              { time: '14:34:55', user: 'system', action: 'Request Blocked', key: 'dev-v1', model: '—', status: 'مسدود', ok: false, ip: '92.x.x.x' },
              { time: '14:34:20', user: 'sara@co', action: 'Policy Updated', key: 'staging', model: 'claude-3', status: 'موفق', ok: true, ip: '185.x.x.x' },
              { time: '14:33:48', user: 'bot-1', action: 'Data Redacted', key: 'prod-v3', model: 'gpt-4o', status: 'Redacted', ok: true, ip: '185.x.x.x' },
            ].map((l, i) => (
              <tr key={i} className="border-b border-slate-800/50 last:border-0">
                <td className="py-2 text-slate-400">{l.time}</td>
                <td className="py-2 text-slate-300 font-medium">{l.user}</td>
                <td className="py-2 text-slate-300">{l.action}</td>
                <td className="py-2 text-slate-400">{l.key}</td>
                <td className="py-2 text-slate-400">{l.model}</td>
                <td className="py-2">
                  <span className={`px-1.5 py-0.5 rounded-full text-[9px] font-bold ${
                    l.status === 'موفق' ? 'bg-emerald-500/20 text-emerald-400'
                    : l.status === 'Redacted' ? 'bg-purple-500/20 text-purple-400'
                    : 'bg-rose-500/20 text-rose-400'
                  }`}>{l.status}</span>
                </td>
                <td className="py-2 text-slate-500">{l.ip}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ─── Policy Matrix ───────────────────────────────────────────────────────────
function PolicyMatrix() {
  const teams = [
    { name: 'Product', models: 'gpt-4o, claude', env: 'همه', limit: '۵k/روز', access: 'Read/Write', status: 'فعال', ok: true },
    { name: 'Backend', models: 'همه', env: 'همه', limit: '۵۰k/روز', access: 'Full', status: 'فعال', ok: true },
    { name: 'Support', models: 'gpt-3.5', env: 'Prod', limit: '۲k/روز', access: 'Limited', status: 'محدود', ok: false },
    { name: 'Finance', models: 'Viewer only', env: '—', limit: '—', access: 'Read', status: 'فعال', ok: true },
  ];
  return (
    <div className="overflow-x-auto rounded-xl border border-slate-800 font-sans">
      <table className="w-full text-[10px] text-right">
        <thead>
          <tr className="bg-slate-800/80 text-slate-400">
            {['تیم', 'مدل‌های مجاز', 'محیط', 'سقف مصرف', 'سطح دسترسی', 'Policy'].map(h => (
              <th key={h} className="px-3 py-2.5 font-medium">{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {teams.map((t, i) => (
            <tr key={i} className="border-t border-slate-800 hover:bg-slate-800/30 transition-colors">
              <td className="px-3 py-2.5 font-bold text-white">{t.name}</td>
              <td className="px-3 py-2.5 text-slate-300">{t.models}</td>
              <td className="px-3 py-2.5 text-slate-400">{t.env}</td>
              <td className="px-3 py-2.5 text-slate-300">{t.limit}</td>
              <td className="px-3 py-2.5 text-slate-300">{t.access}</td>
              <td className="px-3 py-2.5">
                <span className={`px-2 py-0.5 rounded-full font-bold text-[9px] ${
                  t.ok ? 'bg-emerald-500/20 text-emerald-400' : 'bg-amber-500/20 text-amber-400'
                }`}>{t.status}</span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ─── CTA Mini Dashboard ──────────────────────────────────────────────────────
function CTAMiniDashboard() {
  return (
    <div className="bg-white/10 backdrop-blur-sm border border-white/20 rounded-2xl p-5 font-sans text-sm text-white">
      <div className="flex items-center gap-2 mb-4 pb-3 border-b border-white/10">
        <Shield className="w-4 h-4 text-indigo-300" />
        <span className="text-xs font-bold text-white/80">Security & Compliance — وضعیت لحظه‌ای</span>
        <span className="mr-auto flex items-center gap-1 text-[10px] text-emerald-300 bg-emerald-500/20 px-2 py-0.5 rounded-full">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse inline-block" />
          Live
        </span>
      </div>
      <div className="grid grid-cols-2 gap-3">
        {[
          { label: '۲۴ Virtual Key فعال', icon: Key, color: 'text-indigo-300' },
          { label: '۹۸٪ Policy Coverage', icon: ShieldCheck, color: 'text-emerald-300' },
          { label: '۱,۲۴۰ درخواست Redacted', icon: EyeOff, color: 'text-purple-300' },
          { label: '۳۲ درخواست مسدودشده', icon: ShieldAlert, color: 'text-rose-300' },
          { label: '۹۹.۹٪ دسترسی پایدار', icon: CheckCircle, color: 'text-teal-300' },
          { label: 'Audit Logs — همه رویدادها', icon: ClipboardList, color: 'text-amber-300' },
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
    question: 'Security & Compliance گُدار AI دقیقاً چه کاری انجام می‌دهد؟',
    answer: 'گُدار AI یک لایه مرکزی برای مدیریت کلیدها، دسترسی‌ها، Policyها، Audit Logs و محافظت از داده‌های حساس در مسیر تعامل با مدل‌های AI فراهم می‌کند.',
  },
  {
    question: 'Virtual Key چه تفاوتی با API Key اصلی دارد؟',
    answer: 'API Key اصلی مربوط به تأمین‌کننده مدل است، اما Virtual Key در گُدار AI ساخته می‌شود و می‌توان برای آن محدودیت، سطح دسترسی، سقف مصرف، مدل‌های مجاز و وضعیت فعال یا غیرفعال تعریف کرد.',
  },
  {
    question: 'آیا می‌توانم جلوی ارسال داده‌های حساس به مدل را بگیرم؟',
    answer: 'بله. با Data Redaction می‌توانید داده‌هایی مثل شماره تماس، ایمیل، کد ملی، Token، Secret یا Patternهای سفارشی را قبل از ارسال به مدل یا ذخیره در لاگ‌ها حذف یا ماسک کنید.',
  },
  {
    question: 'آیا می‌توانم مشخص کنم چه IPهایی اجازه اتصال دارند؟',
    answer: 'بله. با Inbound Rules می‌توانید دسترسی را بر اساس IP، Geo، محیط، تیم یا سایر Metadataها کنترل کنید.',
  },
  {
    question: 'آیا فعالیت کاربران و تغییرات کلیدها قابل رهگیری است؟',
    answer: 'بله. Audit Logs به شما کمک می‌کند فعالیت‌های مهم مثل ساخت کلید، تغییر دسترسی، درخواست‌های مسدودشده، مصرف مدل‌ها و تغییر Policyها را بررسی کنید.',
  },
  {
    question: 'آیا برای استفاده از گُدار AI نیاز به پرداخت ارزی دارم؟',
    answer: 'خیر. گُدار AI برای تیم‌های ایرانی طراحی شده و هدف آن ساده‌کردن دسترسی به مدل‌های AI بدون نیاز به پرداخت ارزی مستقیم و بدون درگیری با محدودیت‌های رایج سرویس‌های خارجی است.',
  },
];

// ─── Main Component ───────────────────────────────────────────────────────────
export function SecurityCompliance() {
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
              <Shield className="w-4 h-4" />
              Security & Compliance برای تیم‌های ایرانی AI
            </motion.div>

            <motion.h1
              className="text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight mb-8 leading-tight text-slate-900 dark:text-white"
              style={{ fontFamily: 'var(--font-heading)' }}
              {...fadeIn}
            >
              امنیت، کنترل دسترسی و{' '}
              <span className="text-transparent bg-clip-text bg-gradient-to-l from-indigo-500 to-purple-500">
                Compliance
              </span>{' '}
              برای تمام تعاملات LLM
            </motion.h1>

            <motion.p
              className="text-lg md:text-xl text-slate-600 dark:text-slate-400 mb-10 leading-relaxed max-w-3xl mx-auto"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
            >
              با گُدار AI، کلیدهای API، داده‌های حساس، دسترسی تیم‌ها و مسیر درخواست‌های AI را از یک لایه مرکزی مدیریت کنید؛ بدون افشای کلیدها، بدون پراکندگی دسترسی‌ها و بدون درگیری با تحریم یا پرداخت ارزی مستقیم.
            </motion.p>

            <motion.div
              className="flex flex-col sm:flex-row items-center justify-center gap-4"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
            >
              <Link
                href="/signup"
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
              className="absolute inset-0 bg-gradient-to-t from-white via-transparent to-transparent dark:from-[#0B0F19] z-20 pointer-events-none"
              style={{ top: '65%' }}
            />
            <SecurityDashboard />
          </motion.div>
        </div>
      </section>

      {/* ═══ 2. Trust Bar ════════════════════════════════════════════════════ */}
      <section className="py-12 border-y border-slate-100 dark:border-slate-800/50 bg-slate-50/50 dark:bg-slate-900/20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-8">
            <h2 className="text-xl font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
              ساخته‌شده برای تیم‌هایی که AI را امن و قابل‌کنترل وارد محصول می‌کنند
            </h2>
            <p className="text-sm text-slate-600 dark:text-slate-400 max-w-2xl mx-auto">
              گُدار AI به تیم‌های فنی، محصول، امنیت و مدیریت کمک می‌کند استفاده از مدل‌های LLM را با کنترل دسترسی، مدیریت کلیدها، لاگ‌های قابل‌ردیابی و سیاست‌های امنیتی شفاف به سطح Production برسانند.
            </p>
          </div>
          <div className="flex flex-wrap justify-center gap-3">
            {[
              'استارتاپ‌های AI', 'تیم‌های Backend و Platform', 'تیم‌های امنیت و DevOps',
              'محصولات SaaS', 'سازمان‌های داده‌محور', 'تیم‌های Customer Support',
              'تیم‌های R&D و AI Lab', 'شرکت‌هایی که با داده‌های حساس کار می‌کنند',
            ].map((chip, idx) => (
              <div key={idx} className="px-4 py-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-full text-sm font-medium text-slate-700 dark:text-slate-300 shadow-sm">
                {chip}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 3. Stay Compliant, Stay Secure ══════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              امن بمانید، قابل‌ممیزی بمانید
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              گُدار AI کمک می‌کند کلیدها و داده‌های حساس شما در مسیر تعامل با مدل‌های AI افشا نشوند و استفاده از LLMها تحت سیاست‌های مشخص و قابل‌پیگیری انجام شود.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Card 1 — Virtual Keys */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(99,102,241,0.1)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-indigo-100 dark:bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 rounded-xl flex items-center justify-center mb-6">
                  <Key className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  Virtual Keys برای تأمین‌کننده‌های AI
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  به‌جای اینکه کلید اصلی هر تأمین‌کننده را در چند محصول، چند مخزن کد یا چند محیط توسعه پخش کنید، آن را در گُدار AI متصل کنید و برای هر تیم، محصول یا محیط، Virtual Key جداگانه بسازید.
                </p>
                <p className="text-xs text-indigo-600 dark:text-indigo-400 font-medium mb-6">
                  امنیت بیشتر، چرخش ساده‌تر کلیدها، محدودیت‌گذاری دقیق‌تر و رهگیری مصرف در سطح هر کلید.
                </p>
                <VirtualKeysPanel />
              </div>
            </MagicCard>

            {/* Card 2 — Inbound Rules */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(139,92,246,0.1)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-purple-100 dark:bg-purple-500/20 text-purple-600 dark:text-purple-400 rounded-xl flex items-center justify-center mb-6">
                  <Filter className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  کنترل دسترسی و Inbound Rules
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  مشخص کنید چه IPها، کشورها، تیم‌ها یا محیط‌هایی اجازه اتصال به Gateway گُدار AI را دارند. با Access Control و قوانین ورودی، ریسک استفاده غیرمجاز یا اتصال مشکوک را کاهش دهید.
                </p>
                <p className="text-xs text-purple-600 dark:text-purple-400 font-medium mb-6">
                  مناسب برای تیم‌هایی که سیاست‌های امنیتی، محدودیت شبکه یا الزامات داخلی Compliance دارند.
                </p>
                <InboundRulesBuilder />
              </div>
            </MagicCard>

            {/* Card 3 — Data Redaction */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(239,68,68,0.07)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-rose-100 dark:bg-rose-500/20 text-rose-600 dark:text-rose-400 rounded-xl flex items-center justify-center mb-6">
                  <EyeOff className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  حذف خودکار داده‌های حساس با Data Redaction
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  گُدار AI می‌تواند داده‌های حساس را قبل از ارسال درخواست به مدل شناسایی و حذف یا ماسک کند. اطلاعاتی مثل شماره تماس، ایمیل، کد ملی، توکن‌های داخلی، کلیدها یا داده‌های شخصی می‌توانند از Promptها و لاگ‌ها پاک‌سازی شوند.
                </p>
                <p className="text-xs text-rose-600 dark:text-rose-400 font-medium mb-6">
                  داده حساس را قبل از اینکه به مدل یا لاگ برسد کنترل کنید.
                </p>
                <DataRedactionVisual />
              </div>
            </MagicCard>

            {/* Card 4 — Compliance */}
            <MagicCard className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl overflow-hidden" gradientColor="rgba(16,185,129,0.08)">
              <div className="p-8 flex flex-col">
                <div className="w-12 h-12 bg-emerald-100 dark:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 rounded-xl flex items-center justify-center mb-6">
                  <ShieldCheck className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3" style={{ fontFamily: 'var(--font-heading)' }}>
                  Compliance با سیاست‌های امنیتی تیم شما
                </h3>
                <p className="text-slate-600 dark:text-slate-400 mb-4 leading-relaxed flex-1">
                  وقتی AI وارد محصول واقعی می‌شود، تیم‌ها باید بدانند چه کسی، با چه کلیدی، از کدام مدل، در چه زمانی و با چه داده‌ای استفاده کرده است. گُدار AI با لاگ‌های قابل‌ردیابی، سیاست‌های دسترسی و کنترل متمرکز، مسیر استفاده از AI را شفاف‌تر و قابل‌ممیزی‌تر می‌کند.
                </p>
                <p className="text-xs text-emerald-600 dark:text-emerald-400 font-medium mb-6">
                  برای Audit، گزارش‌گیری داخلی، کنترل ریسک و آماده‌سازی برای استانداردهای امنیتی.
                </p>
                <ComplianceStatusCard />
              </div>
            </MagicCard>
          </div>
        </div>
      </section>

      {/* ═══ 4. Security Layer for Production AI ══════════════════════════════ */}
      <section className="py-24 bg-slate-900 dark:bg-[#060913] text-white border-y border-slate-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              یک لایه امنیتی برای AI در مقیاس Production
            </h2>
            <p className="text-lg text-slate-400">
              اتصال مستقیم به مدل‌ها شاید برای Prototype کافی باشد، اما برای محصول واقعی به کنترل کلیدها، قوانین دسترسی، رهگیری مصرف، محدودیت‌گذاری و محافظت از داده نیاز دارید.
            </p>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            {/* Left: Capabilities */}
            <div className="space-y-6">
              {[
                {
                  icon: Key, color: 'text-indigo-400', bg: 'bg-indigo-500/20',
                  title: 'مدیریت متمرکز API Keys',
                  desc: 'کلیدهای اصلی تأمین‌کننده‌ها را در یک نقطه امن مدیریت کنید و به‌جای اشتراک‌گذاری مستقیم آن‌ها، از Virtual Keys قابل‌کنترل در تیم‌ها و سرویس‌های مختلف استفاده کنید.',
                },
                {
                  icon: Settings, color: 'text-amber-400', bg: 'bg-amber-500/20',
                  title: 'محدودیت‌گذاری مصرف در سطح تیم و کلید',
                  desc: 'برای هر تیم، محصول، محیط یا کلید، سقف مصرف، مدل‌های مجاز، Rate Limit و سیاست‌های مشخص تعریف کنید تا هزینه و ریسک امنیتی کنترل شود.',
                },
                {
                  icon: ClipboardList, color: 'text-emerald-400', bg: 'bg-emerald-500/20',
                  title: 'Audit Logs برای رهگیری کامل',
                  desc: 'همه فعالیت‌های مهم مثل ساخت کلید، تغییر دسترسی، مصرف مدل، خطا، درخواست مسدودشده یا تغییر Policy را ثبت کنید تا بعداً قابل بررسی و گزارش‌گیری باشند.',
                },
                {
                  icon: Users, color: 'text-purple-400', bg: 'bg-purple-500/20',
                  title: 'Policyهای قابل تنظیم برای تیم‌های مختلف',
                  desc: 'برای تیم‌های Product، Backend، Support، Finance یا R&D سیاست‌های جداگانه تعریف کنید. هر تیم فقط به مدل‌ها، مسیرها و قابلیت‌هایی دسترسی داشته باشد که واقعاً نیاز دارد.',
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

            {/* Right: Audit Log Table */}
            <div className="relative">
              <div className="absolute -inset-4 bg-gradient-to-br from-indigo-500/10 to-emerald-500/10 rounded-3xl blur-2xl z-0" />
              <div className="relative z-10">
                <AuditLogTable />
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ═══ 5. Protect Sensitive Data ════════════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              داده‌های حساس را قبل از رسیدن به مدل کنترل کنید
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              در AI appها، داده‌های کاربران ممکن است ناخواسته وارد Prompt، Response یا Logs شوند. گُدار AI کمک می‌کند این داده‌ها قبل از افشا شناسایی، ماسک یا حذف شوند.
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
            {[
              {
                icon: EyeOff, color: 'text-rose-600 dark:text-rose-400', bg: 'bg-rose-50 dark:bg-rose-500/10',
                title: 'Redaction برای اطلاعات شخصی',
                desc: 'اطلاعاتی مثل شماره تماس، ایمیل، نام، کد ملی، آدرس یا شناسه‌های داخلی را قبل از ارسال به مدل یا ذخیره در لاگ‌ها شناسایی و پاک‌سازی کنید.',
                tags: ['Phone', 'Email', 'National ID', 'Address'],
              },
              {
                icon: Lock, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10',
                title: 'ماسک کردن Secrets و کلیدها',
                desc: 'اگر کلید API، توکن داخلی، Authorization Header یا Secret به‌اشتباه وارد درخواست شود، گُدار AI می‌تواند آن را شناسایی و ماسک کند.',
                tags: ['API Key', 'Token', 'Secret', 'Auth Header'],
              },
              {
                icon: Settings, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10',
                title: 'الگوهای سفارشی برای داده‌های سازمانی',
                desc: 'تیم شما می‌تواند Patternهای اختصاصی برای داده‌های حساس خود تعریف کند؛ مثل شناسه مشتری، شماره قرارداد، کد داخلی یا ساختارهای خاص کسب‌وکار.',
                tags: ['Customer ID', 'Contract No.', 'Custom Pattern'],
              },
              {
                icon: FileText, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10',
                title: 'کنترل داده در Prompt و Logs',
                desc: 'داده‌های حساس فقط در مسیر ارسال به مدل مهم نیستند؛ باید در لاگ‌ها، گزارش‌ها و داشبوردها هم کنترل شوند تا دسترسی داخلی نیز امن بماند.',
                tags: ['Prompt', 'Response', 'Logs', 'Dashboard'],
              },
            ].map((card, i) => (
              <motion.div
                key={i}
                className="bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 hover:border-indigo-200 dark:hover:border-indigo-800 transition-colors group"
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: (i % 2) * 0.1 }}
              >
                <div className={`w-10 h-10 ${card.bg} ${card.color} rounded-xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform`}>
                  <card.icon className="w-5 h-5" />
                </div>
                <h3 className="font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                  {card.title}
                </h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed mb-4">{card.desc}</p>
                <div className="flex flex-wrap gap-1.5">
                  {card.tags.map(t => (
                    <span key={t} className="px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 text-[10px] font-sans font-medium">{t}</span>
                  ))}
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 6. Governance ════════════════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              Governance برای استفاده تیمی از LLM
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              وقتی چند تیم از AI استفاده می‌کنند، باید بدانید چه کسی به چه چیزی دسترسی دارد، چه مصرفی ایجاد می‌کند و کدام Policy روی درخواست‌ها اعمال می‌شود.
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5 mb-12">
            {[
              {
                icon: Users, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10',
                title: 'دسترسی بر اساس نقش',
                desc: 'سطح دسترسی کاربران را بر اساس نقش‌هایی مثل Admin، Developer، Viewer، Finance یا Security مشخص کنید.',
              },
              {
                icon: Layers, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10',
                title: 'جداسازی محیط‌ها',
                desc: 'کلیدها و Policyهای محیط Development، Staging و Production را جدا نگه دارید تا خطاهای توسعه وارد محصول واقعی نشوند.',
              },
              {
                icon: Eye, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10',
                title: 'دید کامل روی مصرف',
                desc: 'مصرف AI را در سطح کاربر، تیم، کلید، مدل، تأمین‌کننده و محصول مشاهده کنید.',
              },
              {
                icon: AlertTriangle, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10',
                title: 'هشدارهای امنیتی',
                desc: 'برای افزایش غیرعادی مصرف، خطاهای زیاد، تلاش‌های دسترسی ناموفق یا درخواست‌های مشکوک هشدار تعریف کنید.',
              },
              {
                icon: ToggleRight, color: 'text-rose-600 dark:text-rose-400', bg: 'bg-rose-50 dark:bg-rose-500/10',
                title: 'اجرای خودکار Policy',
                desc: 'قوانین امنیتی فقط گزارش نمی‌شوند؛ می‌توانند درخواست را محدود، مسدود، ماسک یا به مسیر امن‌تر هدایت کنند.',
              },
              {
                icon: ClipboardList, color: 'text-teal-600 dark:text-teal-400', bg: 'bg-teal-50 dark:bg-teal-500/10',
                title: 'Audit در سطح سازمان',
                desc: 'تصمیم‌ها، دسترسی‌ها، تغییرات کلید و مصرف مدل‌ها را در سطح کل سازمان قابل‌ردیابی کنید.',
              },
            ].map((item, i) => (
              <motion.div
                key={i}
                className="bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 hover:border-indigo-200 dark:hover:border-indigo-800 transition-colors group"
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: (i % 3) * 0.1 }}
              >
                <div className={`w-10 h-10 ${item.bg} ${item.color} rounded-xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform`}>
                  <item.icon className="w-5 h-5" />
                </div>
                <h3 className="font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>
                  {item.title}
                </h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">{item.desc}</p>
              </motion.div>
            ))}
          </div>

          {/* Policy Matrix */}
          <div className="bg-slate-900 rounded-2xl p-6 overflow-hidden">
            <h3 className="text-sm font-bold text-white mb-4 flex items-center gap-2">
              <Users className="w-4 h-4 text-indigo-400" />
              ماتریس دسترسی تیم‌ها — Policy Matrix
            </h3>
            <PolicyMatrix />
          </div>
        </div>
      </section>

      {/* ═══ 7. Differentiation ══════════════════════════════════════════════ */}
      <section className="py-24 relative">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,#4F46E5_0%,transparent_60%)] opacity-[0.04] dark:opacity-[0.08] pointer-events-none" />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
              چرا Security & Compliance گُدار AI برای تیم‌های ایرانی مهم‌تر است؟
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              برای تیم‌های ایرانی، امنیت AI فقط به محافظت از داده محدود نیست؛ مسئله دسترسی پایدار، پرداخت، ریسک قطع سرویس، کنترل تأمین‌کننده‌ها و مدیریت متمرکز هم مطرح است.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {[
              {
                icon: Shield, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10', border: 'border-indigo-100 dark:border-indigo-500/20',
                title: 'بدون محدودیت‌های تحریمی',
                desc: 'با گُدار AI، تیم شما برای استفاده از مدل‌های AI خارجی کمتر درگیر محدودیت‌های تحریم، مسدود شدن حساب یا قطع دسترسی مستقیم می‌شود.',
              },
              {
                icon: CreditCard, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10', border: 'border-emerald-100 dark:border-emerald-500/20',
                title: 'بدون پرداخت ارزی مستقیم',
                desc: 'به‌جای درگیری با کارت اعتباری خارجی، پرداخت دلاری و حساب‌های بین‌المللی، مصرف AI را از طریق گُدار AI و متناسب با نیاز بازار ایران مدیریت کنید.',
              },
              {
                icon: Lock, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10', border: 'border-purple-100 dark:border-purple-500/20',
                title: 'کنترل متمرکز روی کلیدها و دسترسی‌ها',
                desc: 'گُدار AI کمک می‌کند کلیدها، دسترسی‌ها، مصرف، Policyها و لاگ‌های امنیتی از چند مسیر پراکنده به یک لایه مرکزی منتقل شوند.',
              },
              {
                icon: Globe, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10', border: 'border-amber-100 dark:border-amber-500/20',
                title: 'طراحی‌شده برای تیم‌های فارسی‌زبان',
                desc: 'داشبوردها، گزارش‌ها و تجربه استفاده در گُدار AI برای تیم‌هایی طراحی می‌شود که محصول فارسی، کاربر ایرانی و نیازهای عملیاتی داخل ایران دارند.',
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
              Security & Compliance گُدار AI برای چه سناریوهایی مناسب است؟
            </h2>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            {[
              {
                icon: ShoppingCart, color: 'text-indigo-600 dark:text-indigo-400', bg: 'bg-indigo-50 dark:bg-indigo-500/10',
                title: 'محصولات SaaS با داده کاربران',
                desc: 'اگر کاربران شما اطلاعات شخصی، مکالمه، فایل یا داده حساس وارد محصول می‌کنند، گُدار AI کمک می‌کند مسیر تعامل با مدل‌ها امن‌تر و قابل‌کنترل‌تر شود.',
              },
              {
                icon: Layers, color: 'text-purple-600 dark:text-purple-400', bg: 'bg-purple-50 dark:bg-purple-500/10',
                title: 'تیم‌های Backend و Platform',
                desc: 'مدیریت API Key، محدودیت مصرف، جداسازی محیط‌ها، Rate Limit و قوانین دسترسی را از یک لایه مرکزی کنترل کنید.',
              },
              {
                icon: ShieldAlert, color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-500/10',
                title: 'تیم‌های امنیت و DevOps',
                desc: 'قوانین ورودی، IP Allowlist، Audit Logs، هشدارهای امنیتی و Policyهای اجرایی را برای استفاده امن از AI تعریف کنید.',
              },
              {
                icon: Briefcase, color: 'text-amber-600 dark:text-amber-400', bg: 'bg-amber-50 dark:bg-amber-500/10',
                title: 'تیم‌های مالی و مدیریت',
                desc: 'مصرف AI را در سطح کلید، تیم، مدل یا محصول ببینید و مطمئن شوید استفاده از مدل‌ها تحت محدودیت‌های بودجه و Policy انجام می‌شود.',
              },
              {
                icon: Bot, color: 'text-rose-600 dark:text-rose-400', bg: 'bg-rose-50 dark:bg-rose-500/10',
                title: 'AI Agentها و Workflowهای حساس',
                desc: 'برای Agentها و Workflowهایی که به ابزارها، فایل‌ها یا داده کاربران دسترسی دارند، Policyهای دقیق‌تر و قابل‌ممیزی تعریف کنید.',
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
                title: 'چطور کلیدهای API مدل‌های AI را امن مدیریت کنیم؟',
                desc: 'با Virtual Keys، محدودیت مصرف، چرخش کلید و جداسازی محیط‌ها، ریسک افشای کلیدها را کاهش دهید.',
              },
              {
                tag: 'مقاله', color: 'text-rose-600 dark:text-rose-400', tagBg: 'bg-rose-50 dark:bg-rose-500/10',
                title: 'Data Redaction در AI appها چرا ضروری است؟',
                desc: 'Promptها، Responseها و Logs ممکن است شامل داده حساس باشند. Redaction کمک می‌کند این داده‌ها قبل از افشا کنترل شوند.',
              },
              {
                tag: 'راهنما', color: 'text-emerald-600 dark:text-emerald-400', tagBg: 'bg-emerald-50 dark:bg-emerald-500/10',
                title: 'Governance برای تیم‌هایی که چند مدل LLM استفاده می‌کنند',
                desc: 'وقتی چند تیم، چند مدل و چند تأمین‌کننده درگیر هستند، Policy، Audit و Access Control برای مقیاس‌پذیری ضروری می‌شود.',
              },
            ].map((res, i) => (
              <div
                key={i}
                className="bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 hover:shadow-lg hover:border-slate-300 dark:hover:border-slate-700 transition-all cursor-pointer group flex flex-col"
              >
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
                AI stack خود را امن، قابل‌کنترل و قابل‌ممیزی کنید
              </h2>
              <p className="text-lg text-indigo-200 mb-10 leading-relaxed">
                با گُدار AI، کلیدهای API، Virtual Keys، Access Control، Data Redaction، Audit Logs و Policyهای امنیتی را از یک لایه مرکزی مدیریت کنید؛ بدون محدودیت‌های تحریمی و بدون نیاز به پرداخت ارزی مستقیم.
              </p>
              <div className="flex flex-col sm:flex-row items-start gap-4">
                <Link
                  href="/signup"
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


