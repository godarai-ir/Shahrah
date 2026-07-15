'use client';

import dynamic from 'next/dynamic';
import { Activity, Database, DollarSign, Users } from 'lucide-react';

const RequestsAreaChart = dynamic(
  () => import('./charts/RequestsAreaChart').then((mod) => mod.RequestsAreaChart),
  { ssr: false },
);

export function DashboardPreview() {
  const chartReady = true;

  return (
    <section className="py-14 sm:py-20 lg:py-24 bg-slate-900 text-white overflow-hidden relative">
      {/* Background Glow */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-indigo-500/20 rounded-full blur-[120px] pointer-events-none"></div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <div className="text-center max-w-3xl mx-auto mb-10 sm:mb-16">
          <h2 className="text-2xl sm:text-4xl font-bold mb-4 leading-tight">
            پایش دقیق هر درخواست
          </h2>
          <p className="text-sm sm:text-lg text-gray-400 leading-7 sm:leading-relaxed">
            داشبورد گدارAI گزارش رخدادها، مصرف، هزینه و تاخیر را کنار هم نشان می‌دهد تا تصمیم‌های فنی و مالی قابل بررسی باشند.
          </p>
        </div>

        <div className="bg-[#0B0F19] rounded-2xl border border-gray-800 shadow-2xl p-4 sm:p-6 md:p-8" dir="ltr">
          {/* Top Stats */}
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-3 sm:gap-4 mb-6 sm:mb-8">
            {[
              { label: 'درخواست‌ها', value: '1,234,567', icon: <Activity className="text-blue-400" /> },
              { label: 'مصرف توکن', value: '45.2M', icon: <Database className="text-indigo-400" /> },
              { label: 'هزینه کل', value: '$124.50', icon: <DollarSign className="text-green-400" /> },
              { label: 'کاربران فعال', value: '8,432', icon: <Users className="text-purple-400" /> },
            ].map((stat, i) => (
              <div key={i} className="bg-[#111827] p-4 rounded-xl border border-gray-800 flex items-center justify-between">
                <div>
                  <p className="text-gray-500 text-xs font-medium mb-1">{stat.label}</p>
                  <p className="text-lg sm:text-xl font-bold">{stat.value}</p>
                </div>
                <div className="w-10 h-10 rounded-lg bg-gray-800 flex items-center justify-center">
                  {stat.icon}
                </div>
              </div>
            ))}
          </div>

          {/* Chart Area */}
          <div className="bg-[#111827] p-4 sm:p-6 rounded-xl border border-gray-800">
            <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <h3 className="font-bold">نمای درخواست‌ها</h3>
              <div className="flex gap-2 self-start">
                <span className="px-3 py-1 rounded bg-indigo-500/20 text-indigo-400 text-xs font-medium">۷ روز</span>
                <span className="px-3 py-1 rounded hover:bg-gray-800 text-gray-400 cursor-pointer text-xs font-medium">۳۰ روز</span>
              </div>
            </div>
            
            <div className="h-56 sm:h-64 w-full min-w-0 min-h-[220px] sm:min-h-[250px]">
              {chartReady ? (
                <RequestsAreaChart />
              ) : (
                <div className="h-full w-full animate-pulse rounded-lg bg-[#0f172a]" />
              )}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

