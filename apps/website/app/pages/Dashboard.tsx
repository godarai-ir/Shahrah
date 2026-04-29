'use client';

import { TrendingUp, Activity, DollarSign, Zap } from 'lucide-react';

export function Dashboard() {
  const stats = [
    { icon: Activity, label: 'درخواست‌های امروز', value: '12,543', change: '+12%' },
    { icon: DollarSign, label: 'هزینه ماه جاری', value: '۱۲۳,۴۵۶ تومان', change: '-8%' },
    { icon: Zap, label: 'میانگین تاخیر', value: '۱۲۳ ms', change: '-5%' },
    { icon: TrendingUp, label: 'نرخ موفقیت', value: '۹۹.۸٪', change: '+0.2%' },
  ];

  return (
    <div className="py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">داشبورد</h1>
          <p className="text-slate-600 dark:text-gray-400">
            خوش آمدید! نمای کلی عملکرد API شما
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
          {stats.map((stat) => (
            <div key={stat.label} className="p-6 rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
              <div className="flex items-center justify-between mb-4">
                <stat.icon className="w-8 h-8 text-indigo-500" />
                <span className={`text-sm font-medium ${stat.change.startsWith('+') ? 'text-green-500' : 'text-red-500'}`}>
                  {stat.change}
                </span>
              </div>
              <p className="text-2xl font-bold mb-1">{stat.value}</p>
              <p className="text-sm text-slate-600 dark:text-gray-400">{stat.label}</p>
            </div>
          ))}
        </div>

        <div className="grid lg:grid-cols-2 gap-8">
          <div className="p-8 rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
            <h2 className="text-xl font-bold mb-6">استفاده از API</h2>
            <div className="h-64 flex items-center justify-center text-slate-600 dark:text-gray-400">
              نمودار استفاده روزانه
            </div>
          </div>

          <div className="p-8 rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
            <h2 className="text-xl font-bold mb-6">مدل‌های پرکاربرد</h2>
            <div className="space-y-4">
              {[
                { name: 'GPT-4', usage: 45 },
                { name: 'Claude 3', usage: 30 },
                { name: 'Gemini Pro', usage: 15 },
                { name: 'Llama 3', usage: 10 },
              ].map((model) => (
                <div key={model.name}>
                  <div className="flex justify-between mb-2">
                    <span className="font-medium">{model.name}</span>
                    <span className="text-slate-600 dark:text-gray-400">{model.usage}%</span>
                  </div>
                  <div className="h-2 bg-gray-200 dark:bg-gray-800 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-indigo-500 rounded-full"
                      style={{ width: `${model.usage}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="mt-8 p-8 rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
          <h2 className="text-xl font-bold mb-6">درخواست‌های اخیر</h2>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200 dark:border-gray-800">
                  <th className="text-right py-3 px-4 font-medium">زمان</th>
                  <th className="text-right py-3 px-4 font-medium">مدل</th>
                  <th className="text-right py-3 px-4 font-medium">وضعیت</th>
                  <th className="text-right py-3 px-4 font-medium">تاخیر</th>
                  <th className="text-right py-3 px-4 font-medium">توکن‌ها</th>
                </tr>
              </thead>
              <tbody>
                {[
                  { time: '۱۰:۳۵', model: 'GPT-4', status: 'موفق', latency: '۱۲۳ms', tokens: '۲۵۰' },
                  { time: '۱۰:۳۴', model: 'Claude 3', status: 'موفق', latency: '۹۸ms', tokens: '۱۸۰' },
                  { time: '۱۰:۳۳', model: 'Gemini Pro', status: 'موفق', latency: '۱۴۵ms', tokens: '۳۲۰' },
                  { time: '۱۰:۳۲', model: 'GPT-4', status: 'موفق', latency: '۱۱۰ms', tokens: '۲۹۰' },
                ].map((request, i) => (
                  <tr key={i} className="border-b border-gray-200 dark:border-gray-800">
                    <td className="py-3 px-4 text-slate-600 dark:text-gray-400">{request.time}</td>
                    <td className="py-3 px-4">{request.model}</td>
                    <td className="py-3 px-4">
                      <span className="inline-block px-2 py-1 text-xs rounded-full bg-green-500/10 text-green-500">
                        {request.status}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-slate-600 dark:text-gray-400">{request.latency}</td>
                    <td className="py-3 px-4 text-slate-600 dark:text-gray-400">{request.tokens}</td>
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

