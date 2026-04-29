'use client';

import { motion } from 'motion/react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { Activity, Database, DollarSign, Users } from 'lucide-react';

const data = [
  { name: '1', requests: 4000, cost: 24 },
  { name: '2', requests: 3000, cost: 13 },
  { name: '3', requests: 2000, cost: 98 },
  { name: '4', requests: 2780, cost: 39 },
  { name: '5', requests: 1890, cost: 48 },
  { name: '6', requests: 2390, cost: 38 },
  { name: '7', requests: 3490, cost: 43 },
];

export function DashboardPreview() {
  return (
    <section className="py-24 bg-slate-900 text-white overflow-hidden relative">
      {/* Background Glow */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-indigo-500/20 rounded-full blur-[120px] pointer-events-none"></div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <div className="text-center max-w-3xl mx-auto mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold mb-4">
            نظارت دقیق بر روی هر درخواست
          </h2>
          <p className="text-lg text-gray-400">
            داشبورد حرفه‌ای شاهراه به شما امکان می‌دهد تا تمام معیارها را در لحظه مانیتور کنید و تصمیمات بهتری بگیرید.
          </p>
        </div>

        <div className="bg-[#0B0F19] rounded-2xl border border-gray-800 shadow-2xl p-6 md:p-8" dir="ltr">
          {/* Top Stats */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
            {[
              { label: 'Total Requests', value: '1,234,567', icon: <Activity className="text-blue-400" /> },
              { label: 'Token Usage', value: '45.2M', icon: <Database className="text-indigo-400" /> },
              { label: 'Total Cost', value: '$124.50', icon: <DollarSign className="text-green-400" /> },
              { label: 'Active Users', value: '8,432', icon: <Users className="text-purple-400" /> },
            ].map((stat, i) => (
              <div key={i} className="bg-[#111827] p-4 rounded-xl border border-gray-800 flex items-center justify-between">
                <div>
                  <p className="text-gray-500 text-xs font-medium mb-1">{stat.label}</p>
                  <p className="text-xl font-bold">{stat.value}</p>
                </div>
                <div className="w-10 h-10 rounded-lg bg-gray-800 flex items-center justify-center">
                  {stat.icon}
                </div>
              </div>
            ))}
          </div>

          {/* Chart Area */}
          <div className="bg-[#111827] p-6 rounded-xl border border-gray-800">
            <div className="flex items-center justify-between mb-6">
              <h3 className="font-bold">Requests Overview</h3>
              <div className="flex gap-2">
                <span className="px-3 py-1 rounded bg-indigo-500/20 text-indigo-400 text-xs font-medium">7 Days</span>
                <span className="px-3 py-1 rounded hover:bg-gray-800 text-gray-400 cursor-pointer text-xs font-medium">30 Days</span>
              </div>
            </div>
            
            <div className="h-64 w-full min-w-[200px] min-h-[250px]">
              <ResponsiveContainer width="100%" height="100%" minWidth={200} minHeight={200}>
                <AreaChart data={data}>
                  <defs key="defs">
                    <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#6366F1" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#6366F1" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid key="grid" strokeDasharray="3 3" stroke="#1F2937" vertical={false} />
                  <XAxis key="xaxis" dataKey="name" stroke="#4B5563" tick={{fill: '#9CA3AF', fontSize: 12}} axisLine={false} tickLine={false} />
                  <YAxis key="yaxis" stroke="#4B5563" tick={{fill: '#9CA3AF', fontSize: 12}} axisLine={false} tickLine={false} />
                  <Tooltip key="tooltip" 
                    contentStyle={{ backgroundColor: '#111827', borderColor: '#374151', borderRadius: '8px' }}
                    itemStyle={{ color: '#E5E7EB' }}
                  />
                  <Area key="area" type="monotone" dataKey="requests" stroke="#6366F1" strokeWidth={3} fillOpacity={1} fill="url(#colorRequests)" />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

