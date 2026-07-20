'use client';

import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

const data = [
  { name: '1', requests: 4000, cost: 24 },
  { name: '2', requests: 3000, cost: 13 },
  { name: '3', requests: 2000, cost: 98 },
  { name: '4', requests: 2780, cost: 39 },
  { name: '5', requests: 1890, cost: 48 },
  { name: '6', requests: 2390, cost: 38 },
  { name: '7', requests: 3490, cost: 43 },
];

const INITIAL_DIMENSION = { width: 320, height: 224 } as const;

export function RequestsAreaChart() {
  return (
    <ResponsiveContainer width="100%" height="100%" initialDimension={INITIAL_DIMENSION}>
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
  );
}
