import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Activity, BarChart3, Database, Layers, ArrowUpRight } from 'lucide-react';
import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip } from 'recharts';

interface MetricCardProps {
  title: string;
  value: string;
  change: string;
  icon: React.ReactNode;
}

const MetricCard = ({ title, value, change, icon }: MetricCardProps) => {
  return (
    <motion.div
      whileHover={{ y: -4, transition: { duration: 0.2 } }}
      className="p-6 rounded-2xl bg-slate-900/60 border border-slate-800 backdrop-blur-xl shadow-xl"
    >
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-slate-400">{title}</span>
        <div className="p-2 rounded-lg bg-blue-500/10 text-blue-400">{icon}</div>
      </div>
      <div className="mt-4 flex items-baseline justify-between">
        <h3 className="text-2xl font-bold text-white tracking-tight">{value}</h3>
        <span className="flex items-center text-xs font-semibold text-emerald-400">
          {change} <ArrowUpRight className="w-3 h-3 ml-0.5" />
        </span>
      </div>
    </motion.div>
  );
};

export const App = () => {
  const [data] = useState([
    { time: '00:00', queries: 240, latency: 12 },
    { time: '04:00', queries: 139, latency: 15 },
    { time: '08:00', queries: 980, latency: 18 },
    { time: '12:00', queries: 3908, latency: 22 },
    { time: '16:00', queries: 4800, latency: 24 },
    { time: '20:00', queries: 2400, latency: 16 },
  ]);

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 p-8">
      <header className="mb-8">
        <h1 className="text-3xl font-extrabold tracking-tight bg-gradient-to-r from-blue-400 to-indigo-300 bg-clip-text text-transparent">
          Enterprise Analytics Dashboard
        </h1>
        <p className="text-slate-400 mt-1 text-sm">
          Live telemetry stream & BigQuery conversational analytics.
        </p>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <MetricCard title="Active Agents" value="142" change="+12%" icon={<Activity className="w-5 h-5" />} />
        <MetricCard title="Query Volume" value="1.2M" change="+24%" icon={<Database className="w-5 h-5" />} />
        <MetricCard title="Cache Hit Ratio" value="94.8%" change="+3.2%" icon={<Layers className="w-5 h-5" />} />
      </div>

      <div className="p-6 rounded-2xl bg-slate-900/60 border border-slate-800 backdrop-blur-xl mb-8">
        <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <BarChart3 className="w-5 h-5 text-blue-400" /> Query Throughput Over Time
        </h2>
        <div className="h-72 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={data}>
              <defs>
                <linearGradient id="colorQueries" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.4} />
                  <stop offset="95%" stopColor="#3b82f6" stopOpacity={0.0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="time" stroke="#64748b" />
              <YAxis stroke="#64748b" />
              <Tooltip contentStyle={{ backgroundColor: '#0f172a', borderColor: '#1e293b' }} />
              <Area type="monotone" dataKey="queries" stroke="#3b82f6" fillOpacity={1} fill="url(#colorQueries)" />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
};
