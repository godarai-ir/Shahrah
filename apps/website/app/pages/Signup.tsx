'use client';

import React, { useEffect, useState } from 'react';
import dynamic from 'next/dynamic';
import { motion, AnimatePresence } from 'motion/react';
import { CheckCircle2, Users, Building2, Phone, LayoutGrid, Sparkles, Activity, ArrowRight, ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { MagicCard } from '../components/ui/MagicCard';

type Step = 'contact' | 'team' | 'usage' | 'features' | 'success';

interface FormData {
  mobile: string;
  teamSize: string;
  usageType: string;
  activityField: string;
  features: string[];
}

const TokenAreaChart = dynamic(
  () => import('../components/charts/TokenAreaChart').then((mod) => mod.TokenAreaChart),
  { ssr: false },
);

function DashboardMockup() {
  const [chartReady, setChartReady] = useState(false);

  useEffect(() => {
    setChartReady(true);
  }, []);

  return (
    <div className="w-full max-w-2xl bg-slate-900/90 border border-indigo-500/20 rounded-2xl p-6 shadow-2xl relative overflow-hidden ring-1 ring-white/5 backdrop-blur-xl">
      <div className="absolute -top-32 -right-32 w-64 h-64 bg-indigo-500/30 rounded-full blur-[100px]" />
      <div className="absolute -bottom-32 -left-32 w-64 h-64 bg-purple-500/20 rounded-full blur-[100px]" />
      
      <div className="flex items-center justify-between mb-8 relative z-10">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-indigo-500/10 flex items-center justify-center border border-indigo-500/20 backdrop-blur-sm">
            <Activity className="w-5 h-5 text-indigo-400" />
          </div>
          <div>
            <h3 className="text-white font-bold text-base" style={{ fontFamily: 'var(--font-heading)' }}>مانیتورینگ لحظه‌ای</h3>
            <p className="text-slate-400 text-xs mt-0.5">وضعیت زیرساخت شاهراه</p>
          </div>
        </div>
        <div className="flex items-center gap-2 bg-indigo-500/10 px-3 py-1.5 rounded-full border border-indigo-500/20 backdrop-blur-sm">
          <span className="flex h-2 w-2 relative">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
          </span>
          <span className="text-[11px] text-slate-300 font-medium tracking-wide">همه سیستم‌ها عملیاتی</span>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4 mb-6 relative z-10">
        <MagicCard 
          gradientColor="rgba(99,102,241,0.15)"
          className="bg-slate-800/50 backdrop-blur-md border border-white/5 rounded-xl transition-all hover:bg-slate-800/80 hover:border-indigo-500/30 group"
        >
          <div className="p-4 pointer-events-none w-full h-full">
            <div className="text-slate-400 text-[11px] mb-2 font-medium">مصرف توکن (۲۴ ساعت)</div>
            <div className="text-2xl font-bold text-white font-sans tracking-tight">۲۴.۵M</div>
            <div className="text-emerald-400 text-[10px] mt-2 flex items-center gap-1 font-medium bg-emerald-400/10 w-fit px-2 py-0.5 rounded-full">
              +۱۲٪ نسبت به دیروز
            </div>
          </div>
        </MagicCard>
        <MagicCard 
          gradientColor="rgba(99,102,241,0.15)"
          className="bg-slate-800/50 backdrop-blur-md border border-white/5 rounded-xl transition-all hover:bg-slate-800/80 hover:border-indigo-500/30 group"
        >
          <div className="p-4 pointer-events-none w-full h-full">
            <div className="text-slate-400 text-[11px] mb-2 font-medium">میانگین تاخیر (Latency)</div>
            <div className="text-2xl font-bold text-white font-sans tracking-tight">124ms</div>
            <div className="text-emerald-400 text-[10px] mt-2 flex items-center gap-1 font-medium bg-emerald-400/10 w-fit px-2 py-0.5 rounded-full">
              پاسخ‌دهی بهینه
            </div>
          </div>
        </MagicCard>
        <MagicCard 
          gradientColor="rgba(99,102,241,0.15)"
          className="bg-slate-800/50 backdrop-blur-md border border-white/5 rounded-xl transition-all hover:bg-slate-800/80 hover:border-indigo-500/30 group"
        >
          <div className="p-4 pointer-events-none w-full h-full">
            <div className="text-slate-400 text-[11px] mb-2 font-medium">مدل‌های فعال</div>
            <div className="text-2xl font-bold text-white font-sans tracking-tight">۱۶</div>
            <div className="text-slate-300 text-[10px] mt-2 flex items-center gap-1 font-medium bg-white/5 w-fit px-2 py-0.5 rounded-full">
              از ۳ ارائه‌دهنده
            </div>
          </div>
        </MagicCard>
      </div>

      <div className="h-[200px] w-full min-w-[200px] min-h-[200px] bg-slate-800/50 backdrop-blur-md border border-white/5 rounded-xl p-4 relative z-10">
        {chartReady ? (
          <TokenAreaChart />
        ) : (
          <div className="h-full w-full animate-pulse rounded-lg bg-slate-800/70" />
        )}
      </div>
    </div>
  );
}

export function Signup() {
  const [step, setStep] = useState<Step>('contact');
  const [formData, setFormData] = useState<FormData>({
    mobile: '',
    teamSize: '',
    usageType: '',
    activityField: '',
    features: [],
  });

  const [error, setError] = useState('');
  const [submitError, setSubmitError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const roleOptions = [
    'مدیر محصول',
    'مدیر فنی',
    'معمار نرم‌افزار',
    'توسعه‌دهنده نرم‌افزار',
    'DevOps / MLOps',
    'مدیر عامل'
  ];

  const featureOptions = [
    { name: 'Observability', description: 'ثبت لاگ، ردیابی و پایش متریک‌های کلیدی' },
    { name: 'AI Gateway', description: 'مسیریابی به بیش از ۱۶۰۰ مدل زبانی بزرگ' },
    { name: 'Guardrails', description: 'اعمال کنترل و محدودیت‌های بلادرنگ روی رفتار مدل‌های زبانی' },
    { name: 'Prompt Engineering', description: 'ساخت، نسخه‌بندی و استقرار پرامپت‌ها' },
    { name: 'Agents', description: 'استقرار ایجنت‌های هوش مصنوعی در محیط عملیاتی' },
    { name: 'Security and Compliance', description: 'امنیت در سطح سازمانی' },
    { name: 'MCP Gateway', description: 'استقرار و مدیریت سرورهای MCP در محیط عملیاتی' },
  ];

  const handleNext = async () => {
    if (step === 'contact' && !formData.mobile) {
      setError('لطفاً شماره موبایل خود را وارد کنید');
      return;
    }
    if (step === 'contact' && formData.mobile.length < 10) {
      setError('شماره موبایل نامعتبر است');
      return;
    }
    setError('');

    if (step === 'contact') setStep('team');
    else if (step === 'team') setStep('usage');
    else if (step === 'usage') setStep('features');
    else if (step === 'features') {
      try {
        setIsSubmitting(true);
        setSubmitError('');

        const res = await fetch('/api/signup', {
          method: 'POST',
          headers: { 'content-type': 'application/json' },
          body: JSON.stringify(formData),
        });

        if (!res.ok) {
          throw new Error('request_failed');
        }

        setStep('success');
      } catch {
        setSubmitError('خطا در ذخیره اطلاعات. لطفاً دوباره تلاش کنید.');
      } finally {
        setIsSubmitting(false);
      }
    }
  };

  const handleBack = () => {
    if (step === 'team') setStep('contact');
    else if (step === 'usage') setStep('team');
    else if (step === 'features') setStep('usage');
  };

  const toggleFeature = (featureName: string) => {
    setFormData(prev => ({
      ...prev,
      features: prev.features.includes(featureName)
        ? prev.features.filter(f => f !== featureName)
        : [...prev.features, featureName],
    }));
  };

  const renderProgress = () => {
    const steps: Step[] = ['contact', 'team', 'usage', 'features'];
    const currentIndex = steps.indexOf(step as Step);
    if (currentIndex === -1) return null;

    return (
      <div className="flex gap-1.5 mb-10 w-full" dir="rtl">
        {steps.map((_, i) => (
          <div
            key={i}
            className={`h-1.5 flex-1 rounded-full transition-all duration-300 ${
              i <= currentIndex ? 'bg-indigo-600 dark:bg-indigo-500 shadow-[0_0_10px_rgba(99,102,241,0.5)]' : 'bg-slate-200 dark:bg-slate-800'
            }`}
          />
        ))}
      </div>
    );
  };

  return (
    <div className="flex flex-col lg:flex-row-reverse min-h-screen bg-white dark:bg-[#0B0F19] w-full font-sans text-slate-900 dark:text-white selection:bg-indigo-500/30" dir="rtl">
      
      {/* Form Panel (Left side) */}
      <div className="w-full lg:w-1/2 flex flex-col p-6 lg:p-12 bg-white dark:bg-[#0B0F19] min-h-screen relative z-10 overflow-y-auto border-l lg:border-l-0 lg:border-r border-slate-200 dark:border-slate-800/50">
        <div className="w-full max-w-md mx-auto flex-1 flex flex-col justify-center">
          
          <div className="mb-12">
            <div className="hidden">
              {/* Logo moved to visual panel */}
            </div>
          </div>

          <div className="flex-1">
            <AnimatePresence mode="wait">
              {/* STEP 1: CONTACT */}
              {step === 'contact' && (
                <motion.div
                  key="contact"
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  transition={{ duration: 0.3 }}
                >
                  {renderProgress()}
                  <div className="inline-flex items-center gap-2 px-3 py-1 bg-indigo-50 dark:bg-indigo-500/10 rounded-full mb-6 border border-indigo-100 dark:border-indigo-500/20">
                    <Sparkles className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
                    <span className="text-xs font-bold text-indigo-600 dark:text-indigo-400" style={{ fontFamily: 'var(--font-heading)' }}>ثبت‌نام زودهنگام (Early Access)</span>
                  </div>
                  
                  <h2 className="text-3xl font-bold text-slate-900 dark:text-white mb-4 tracking-tight" style={{ fontFamily: 'var(--font-heading)' }}>
                    به شاهراه بپیوندید
                  </h2>
                  <p className="text-slate-500 dark:text-slate-400 mb-10 text-sm leading-relaxed">
                    با عضویت در لیست انتظار، علاوه بر امکانات رایگان شاهراه، <span className="font-bold text-indigo-600 dark:text-indigo-400 border-b border-indigo-200 dark:border-indigo-500/30 pb-0.5">۱۰۰ هزار توکن رایگان</span> هدیه بگیرید.
                  </p>
                  
                  <div className="mb-8">
                    <label className="block text-sm font-bold mb-3 text-slate-900 dark:text-slate-200" style={{ fontFamily: 'var(--font-heading)' }}>شماره موبایل</label>
                    <div className="relative">
                      <div className="absolute inset-y-0 left-0 flex items-center pl-4 pointer-events-none">
                        <Phone className="w-5 h-5 text-slate-400 dark:text-slate-500" />
                      </div>
                      <input
                        type="tel"
                        autoFocus
                        maxLength={11}
                        value={formData.mobile}
                        onChange={(e) => {
                          const val = e.target.value.replace(/[^0-9]/g, '');
                          setFormData({ ...formData, mobile: val });
                          if (error) setError('');
                        }}
                        placeholder="۰۹۱۲۳۴۵۶۷۸۹"
                        className={`w-full p-4 pl-12 border-2 rounded-xl bg-slate-50 dark:bg-slate-900/50 text-left dir-ltr text-slate-900 dark:text-white font-sans focus:outline-none focus:ring-0 focus:border-indigo-500 focus:bg-white dark:focus:bg-slate-900 hover:border-slate-300 dark:hover:border-slate-700 transition-all text-lg ${
                          error ? 'border-red-500/50 bg-red-50/20 dark:bg-red-500/10' : 'border-slate-200 dark:border-slate-800'
                        }`}
                      />
                    </div>
                    {error && (
                      <motion.p initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-red-500 text-xs mt-3 font-bold flex items-center gap-1">
                        <span className="w-1 h-1 rounded-full bg-red-500 inline-block"></span>
                        {error}
                      </motion.p>
                    )}
                    <p className="text-xs text-slate-500 dark:text-slate-500 mt-4 leading-relaxed">شماره شما صرفاً برای اطلاع‌رسانی زمان بازگشایی سرویس و تخصیص هدیه استفاده خواهد شد.</p>
                  </div>

                  <button 
                    onClick={handleNext} 
                    className="w-full py-4 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-sm shadow-[0_0_20px_rgba(79,70,229,0.3)] hover:shadow-[0_0_30px_rgba(79,70,229,0.5)] transition-all active:scale-[0.98] flex items-center justify-center gap-2 group"
                  >ادامه<ArrowLeft className="w-4 h-4 group-hover:-translate-x-1 transition-transform" /></button>
                </motion.div>
              )}

              {/* STEP 2: TEAM INFO */}
              {step === 'team' && (
                <motion.div
                  key="team"
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -20 }}
                  transition={{ duration: 0.3 }}
                >
                  {renderProgress()}
                  
                  <div className="w-12 h-12 bg-indigo-50 dark:bg-indigo-500/10 border border-indigo-100 dark:border-indigo-500/20 rounded-2xl flex items-center justify-center mb-6">
                    <Users className="w-6 h-6 text-indigo-600 dark:text-indigo-400" />
                  </div>
                  
                  <h2 className="text-2xl font-bold mb-3 text-slate-900 dark:text-white tracking-tight" style={{ fontFamily: 'var(--font-heading)' }}>
                    اندازه تیم شما چقدر است؟
                  </h2>
                  <p className="text-slate-500 dark:text-slate-400 mb-8 text-sm leading-relaxed">این اطلاعات به ما کمک می‌کند تا زیرساخت و ظرفیت مناسبی را برای سازمان شما در نظر بگیریم.</p>

                  <div className="grid grid-cols-2 gap-3 mb-10">
                    {['۱-۵ نفر', '۶-۲۰ نفر', '۲۰-۵۰ نفر', '+۵۰ نفر'].map(size => (
                      <MagicCard
                        key={size}
                        gradientColor="rgba(99,102,241,0.15)"
                        onClick={() => setFormData({ ...formData, teamSize: size })}
                        className={`text-center border-2 transition-all rounded-xl cursor-pointer ${
                          formData.teamSize === size 
                            ? 'border-indigo-600 bg-indigo-50 dark:bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 shadow-sm' 
                            : 'border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900/50 text-slate-600 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-900'
                        }`}
                        role="button"
                        tabIndex={0}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault();
                            setFormData({ ...formData, teamSize: size });
                          }
                        }}
                      >
                        <div className="p-4 flex flex-col items-center justify-center gap-3 pointer-events-none w-full h-full">
                          <span className="text-sm font-bold">{size}</span>
                        </div>
                      </MagicCard>
                    ))}
                  </div>

                  <div className="flex flex-col sm:flex-row gap-3">
                    <button 
                      onClick={handleBack} 
                      className="px-6 py-4 border-2 border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 rounded-xl font-bold text-sm hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
                    >
                      بازگشت
                    </button>
                    <button 
                      onClick={handleNext} 
                      disabled={!formData.teamSize}
                      className={`flex-1 py-4 rounded-xl font-bold text-sm transition-all flex items-center justify-center gap-2 group ${
                        formData.teamSize 
                          ? 'bg-indigo-600 text-white hover:bg-indigo-700 shadow-[0_0_20px_rgba(79,70,229,0.3)] active:scale-[0.98]' 
                          : 'bg-slate-100 dark:bg-slate-800 text-slate-400 dark:text-slate-600 cursor-not-allowed'
                      }`}
                    >
                      مرحله بعد
                      <ArrowLeft className="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
                    </button>
                  </div>
                  {submitError && (
                    <p className="mt-3 text-sm text-red-500 font-bold">{submitError}</p>
                  )}
                </motion.div>
              )}

              {/* STEP 3: ROLE TYPE */}
              {step === 'usage' && (
                <motion.div
                  key="usage"
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -20 }}
                  transition={{ duration: 0.3 }}
                >
                  {renderProgress()}
                  
                  <div className="w-12 h-12 bg-indigo-50 dark:bg-indigo-500/10 border border-indigo-100 dark:border-indigo-500/20 rounded-2xl flex items-center justify-center mb-6">
                    <Building2 className="w-6 h-6 text-indigo-600 dark:text-indigo-400" />
                  </div>

                  <h2 className="text-2xl font-bold mb-3 text-slate-900 dark:text-white tracking-tight" style={{ fontFamily: 'var(--font-heading)' }}>
                    نقش شما در تیم چیست؟
                  </h2>
                  <p className="text-slate-500 dark:text-slate-400 mb-8 text-sm leading-relaxed">برای شخصی‌سازی داشبورد و ارائه مستندات مرتبط با تخصص شما.</p>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-10">
                    {roleOptions.map(option => (
                      <MagicCard
                        key={option}
                        gradientColor="rgba(99,102,241,0.15)"
                        onClick={() => setFormData({ ...formData, usageType: option })}
                        className={`text-right border-2 transition-all rounded-xl cursor-pointer ${
                          formData.usageType === option 
                            ? 'border-indigo-600 bg-indigo-50 dark:bg-indigo-500/10 text-indigo-700 dark:text-indigo-300' 
                            : 'border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900/50 text-slate-600 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-700'
                        }`}
                        role="button"
                        tabIndex={0}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault();
                            setFormData({ ...formData, usageType: option });
                          }
                        }}
                      >
                        <div className="p-4 flex items-center justify-between pointer-events-none w-full h-full">
                          <span className="text-sm font-bold">{option}</span>
                          <div className={`w-4 h-4 rounded-full border-2 flex items-center justify-center flex-shrink-0 ${
                            formData.usageType === option ? 'border-indigo-600 dark:border-indigo-400' : 'border-slate-300 dark:border-slate-600'
                          }`}>
                            {formData.usageType === option && <div className="w-2 h-2 bg-indigo-600 dark:bg-indigo-400 rounded-full" />}
                          </div>
                        </div>
                      </MagicCard>
                    ))}
                  </div>

                  <div className="flex flex-col sm:flex-row gap-3">
                    <button 
                      onClick={handleBack} 
                      className="px-6 py-4 border-2 border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 rounded-xl font-bold text-sm hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
                    >
                      بازگشت
                    </button>
                    <button 
                      onClick={handleNext} 
                      disabled={!formData.usageType}
                      className={`flex-1 py-4 rounded-xl font-bold text-sm transition-all flex items-center justify-center gap-2 group ${
                        formData.usageType 
                          ? 'bg-indigo-600 text-white hover:bg-indigo-700 shadow-[0_0_20px_rgba(79,70,229,0.3)] active:scale-[0.98]' 
                          : 'bg-slate-100 dark:bg-slate-800 text-slate-400 dark:text-slate-600 cursor-not-allowed'
                      }`}
                    >
                      مرحله بعد
                      <ArrowLeft className="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
                    </button>
                  </div>
                </motion.div>
              )}

              {/* STEP 4: FEATURES */}
              {step === 'features' && (
                <motion.div
                  key="features"
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -20 }}
                  transition={{ duration: 0.3 }}
                >
                  {renderProgress()}
                  
                  

                  <h2 className="text-2xl font-bold mb-3 text-slate-900 dark:text-white tracking-tight" style={{ fontFamily: 'var(--font-heading)' }}>
                    امکانات مورد نیاز شما
                  </h2>
                  <p className="text-slate-500 dark:text-slate-400 mb-8 text-sm leading-relaxed">کدام یک از قابلیت‌های پیشرفته شاهراه برای محصول شما اولویت بالاتری دارد؟ </p>

                                                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-10 w-full">
                    {featureOptions.map(option => (
                      <MagicCard
                        key={option.name}
                        gradientColor="rgba(99,102,241,0.15)"
                        className={`text-right border-2 transition-all rounded-xl relative group h-[100px] cursor-pointer ${
                          formData.features.includes(option.name) 
                            ? 'border-indigo-600 bg-indigo-50/50 dark:bg-indigo-900/20 shadow-[0_0_15px_rgba(79,70,229,0.15)] scale-[0.98]' 
                            : 'border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900/50 hover:border-indigo-300 dark:hover:border-indigo-700/50 hover:shadow-md'
                        }`}
                        onClick={() => toggleFeature(option.name)}
                        role="button"
                        tabIndex={0}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault();
                            toggleFeature(option.name);
                          }
                        }}
                      >
                        <div className="flex flex-col justify-start h-full p-4 pointer-events-none">
                          {/* Background subtle gradient for selected state */}
                          {formData.features.includes(option.name) && (
                            <div className="absolute inset-0 bg-gradient-to-br from-indigo-500/5 to-purple-500/5 pointer-events-none" />
                          )}

                          <div className="flex items-center gap-3 z-10 w-full mb-1.5">
                            <div className={`w-5 h-5 rounded-md border-2 flex items-center justify-center flex-shrink-0 transition-all ${
                              formData.features.includes(option.name) 
                                ? 'border-indigo-600 bg-indigo-600 dark:border-indigo-500 dark:bg-indigo-500 shadow-[0_0_10px_rgba(79,70,229,0.3)]' 
                                : 'border-slate-300 dark:border-slate-600 bg-transparent group-hover:border-indigo-400 dark:group-hover:border-indigo-500'
                            }`}>
                              {formData.features.includes(option.name) && <CheckCircle2 className="w-3.5 h-3.5 text-white" />}
                            </div>
                            
                            <span className={`text-sm font-bold dir-ltr text-right line-clamp-1 flex-1 ${
                              formData.features.includes(option.name) ? 'text-indigo-700 dark:text-indigo-300' : 'text-slate-800 dark:text-slate-300 group-hover:text-indigo-600 dark:group-hover:text-indigo-400'
                            }`}>{option.name}</span>
                          </div>

                          <div className="z-10 w-full pl-8 pr-1">
                            <span className={`text-[12px] leading-snug line-clamp-2 ${
                              formData.features.includes(option.name) ? 'text-indigo-600/80 dark:text-indigo-400/80' : 'text-slate-500 dark:text-slate-500'
                            }`}>{option.description}</span>
                          </div>
                        </div>
                      </MagicCard>
                    ))}
                  </div>

                  <div className="flex flex-col sm:flex-row gap-3">
                    <button 
                      onClick={handleBack} 
                      className="px-6 py-4 border-2 border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 rounded-xl font-bold text-sm hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
                    >
                      بازگشت
                    </button>
                    <button 
                      onClick={handleNext} 
                      disabled={formData.features.length === 0 || isSubmitting}
                      className={`flex-1 py-4 rounded-xl font-bold text-sm transition-all flex items-center justify-center gap-2 ${
                        formData.features.length > 0 && !isSubmitting
                          ? 'bg-indigo-600 text-white hover:bg-indigo-700 shadow-[0_0_20px_rgba(79,70,229,0.3)] active:scale-[0.98]' 
                          : 'bg-slate-100 dark:bg-slate-800 text-slate-400 dark:text-slate-600 cursor-not-allowed'
                      }`}
                    >
                      تکمیل ثبت‌نام
                    </button>
                  </div>
                </motion.div>
              )}

              {/* STEP 5: SUCCESS */}
              {step === 'success' && (
                <motion.div
                  key="success"
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ duration: 0.4, type: 'spring' }}
                  className="text-center py-12 flex flex-col items-center"
                >
                  <motion.div 
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    transition={{ delay: 0.2, type: 'spring' }}
                    className="w-24 h-24 bg-green-50 dark:bg-green-500/10 rounded-full flex items-center justify-center mb-8 relative"
                  >
                    <div className="absolute inset-0 bg-green-500 rounded-full animate-ping opacity-20"></div>
                    <CheckCircle2 className="w-12 h-12 text-green-500 dark:text-green-400 relative z-10" />
                  </motion.div>
                  
                  <h2 className="text-3xl font-bold mb-4 text-slate-900 dark:text-white tracking-tight" style={{ fontFamily: 'var(--font-heading)' }}>
                    عضویت با موفقیت انجام شد
                  </h2>
                  <p className="text-slate-600 dark:text-slate-400 mb-10 leading-relaxed max-w-sm mx-auto">
                    شما در لیست انتظار قرار گرفتید. <span className="font-bold text-indigo-600 dark:text-indigo-400 border-b border-indigo-200 dark:border-indigo-500/30 pb-0.5">۱۰۰ هزار توکن هدیه</span> شما در اولین ورود به پنل فعال خواهد شد. به زودی از طریق پیامک به شما اطلاع می‌دهیم.
                  </p>
                  
                  <Link
                    href="/"
                    className="w-full max-w-xs py-4 bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-xl shadow-[0_0_20px_rgba(79,70,229,0.3)] transition-all active:scale-[0.98] inline-block"
                  >
                    بازگشت به صفحه اصلی
                  </Link>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
          
          {/* Form Footer */}
          <div className="mt-8 pt-8 border-t border-slate-100 dark:border-slate-800 text-center">
            <p className="text-xs text-slate-400 dark:text-slate-500">
              با ثبت‌نام در شاهراه، با <Link href="/terms" className="text-indigo-600 dark:text-indigo-400 hover:underline font-bold">قوانین و مقررات</Link> و <Link href="/privacy" className="text-indigo-600 dark:text-indigo-400 hover:underline font-bold">حریم خصوصی</Link> ما موافقت می‌کنید.
            </p>
          </div>

        </div>
      </div>

      {/* Visuals Panel (Left side) */}
      <div className="hidden lg:flex w-full lg:w-1/2 bg-[#0B0F19] p-12 items-center justify-center relative overflow-hidden flex-col">
        
        {/* Decorative Background */}
        <div className="absolute inset-0 opacity-20 pointer-events-none" style={{ backgroundImage: 'radial-gradient(circle at 2px 2px, rgba(255,255,255,0.1) 1px, transparent 0)', backgroundSize: '32px 32px' }}></div>
        <div className="absolute top-0 right-0 w-full h-full bg-gradient-to-b from-[#0B0F19] via-transparent to-[#0B0F19] pointer-events-none z-0"></div>
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[500px] opacity-20 pointer-events-none">
          <div className="absolute inset-0 bg-gradient-to-r from-indigo-500 to-purple-500 blur-[120px] rounded-full mix-blend-screen" />
        </div>

        <div className="relative z-10 w-full flex flex-col items-center h-full">
          
          {/* Header Row: Logo on right, Tag on left */}
          <motion.div 
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="w-full flex justify-between items-center px-8 mb-auto mt-0"
          >
            <Link href="/" className="inline-flex items-center gap-3 text-white hover:opacity-80 transition-opacity">
              <div className="w-12 h-12 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/20">
                <span className="text-white font-bold text-2xl" style={{ fontFamily: 'var(--font-heading)' }}>ش</span>
              </div>
              <span className="font-bold text-3xl tracking-tight" style={{ fontFamily: 'var(--font-heading)' }}>شاهراه</span>
            </Link>

            <div className="inline-flex items-center gap-2 px-3 py-1.5 bg-indigo-500/10 border border-indigo-500/20 rounded-full backdrop-blur-md">
              <span className="flex h-2 w-2 relative">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-indigo-500"></span>
              </span>
              <span className="text-xs font-bold text-indigo-300 tracking-widest uppercase dir-ltr">Developer Preview</span>
            </div>
          </motion.div>

          <motion.div 
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
            className="text-center mb-16 max-w-lg mt-12"
          >
            <p className="text-slate-400 leading-relaxed font-medium text-[16px]">
              شاهراه زیرساخت مقیاس‌پذیر، امن و پرسرعت برای توسعه‌دهندگان ایرانی فراهم می‌کند تا بدون دغدغه تحریم و پرداخت ارزی، محصولات مبتنی بر هوش مصنوعی بسازند.
            </p>
          </motion.div>
          
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.4, type: 'spring' }}
            className="w-full flex justify-center mb-auto"
          >
            <DashboardMockup />
          </motion.div>
        </div>
      </div>
      
    </div>
  );
}

