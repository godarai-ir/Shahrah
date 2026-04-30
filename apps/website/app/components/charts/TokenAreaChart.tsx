'use client';

import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';

const data = [
  { time: '00:00', tokens: 1200 },
  { time: '04:00', tokens: 3000 },
  { time: '08:00', tokens: 2000 },
  { time: '12:00', tokens: 8000 },
  { time: '16:00', tokens: 5500 },
  { time: '20:00', tokens: 9000 },
  { time: '24:00', tokens: 4000 },
];

export function TokenAreaChart() {
  return (
    <ResponsiveContainer width="100%" height={200} minWidth={200}>
      <AreaChart data={data} margin={{ top: 5, right: 0, left: -20, bottom: 0 }}>
        <defs key="defs">
          <linearGradient id="colorTokens" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3}/>
            <stop offset="95%" stopColor="#6366f1" stopOpacity={0}/>
          </linearGradient>
        </defs>
        <CartesianGrid key="grid" strokeDasharray="3 3" stroke="#334155" vertical={false} />
        <XAxis key="xaxis" dataKey="time" stroke="#94a3b8" fontSize={10} tickLine={false} axisLine={false} dy={10} />
        <YAxis key="yaxis" stroke="#94a3b8" fontSize={10} tickLine={false} axisLine={false} tickFormatter={(value) => `${value / 1000}k`} dx={-10} />
        <Tooltip key="tooltip"
          contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px', fontSize: '12px' }}
          itemStyle={{ color: '#fff', fontWeight: 'bold' }}
          labelStyle={{ color: '#94a3b8', marginBottom: '4px' }}
        />
        <Area key="area" type="monotone" dataKey="tokens" stroke="#818cf8" strokeWidth={2} fillOpacity={1} fill="url(#colorTokens)" />
      </AreaChart>
    </ResponsiveContainer>
  );
}
