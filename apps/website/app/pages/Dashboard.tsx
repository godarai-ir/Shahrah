'use client';

import { useEffect, useState } from 'react';
import { TrendingUp, Activity, DollarSign, Zap, Loader2 } from 'lucide-react';

interface Transaction {
  id: string;
  tx_type: string;
  amount_irr: number;
  amount_usd: number;
  status: string;
  description: string;
  created_at: string;
}

interface Wallet {
  balance_irr: number;
  balance_usd: number;
}

export function Dashboard() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [wallet, setWallet] = useState<Wallet | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchData() {
      try {
        // Fetch wallet data via proxy
        const walletRes = await fetch('/api/proxy/wallet?path=/wallet', {
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
        });
        
        // Fetch transactions via proxy
        const txRes = await fetch('/api/proxy/wallet?path=/wallet/transactions?limit=20', {
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
        });

        if (!walletRes.ok || !txRes.ok) {
          throw new Error('Failed to fetch wallet data');
        }

        const walletData = await walletRes.json();
        const txData = await txRes.json();

        setWallet(walletData.data);
        setTransactions(txData.data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
        // Fall back to mock data on error
        setTransactions([
          { id: '1', tx_type: 'CONSUMPTION', amount_irr: 150000, amount_usd: 3.50, status: 'COMPLETED', description: 'inference usage', created_at: new Date().toISOString() },
        ]);
      } finally {
        setLoading(false);
      }
    }

    fetchData();
  }, []);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('fa-IR').format(amount) + ' تومان';
  };

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
  };

  const stats = [
    { icon: Activity, label: 'درخواست‌های امروز', value: '12,543', change: '+12%' },
    { icon: DollarSign, label: 'موجودی کیف پول', value: wallet ? formatCurrency(wallet.balance_irr) : '۱۲۳,۴۵۶ تومان', change: '-8%' },
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
          <h2 className="text-xl font-bold mb-6">تراکنش‌های اخیر</h2>
          
          {loading ? (
            <div className="flex items-center justify-center h-32 text-slate-600 dark:text-gray-400">
              <Loader2 className="w-6 h-6 animate-spin ml-2" />
              در حال بارگذاری...
            </div>
          ) : error ? (
            <div className="flex items-center justify-center h-32 text-amber-600">
              {error}
            </div>
          ) : transactions.length === 0 ? (
            <div className="flex items-center justify-center h-32 text-slate-600 dark:text-gray-400">
              تراکنشی یافت نشد
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-200 dark:border-gray-800">
                    <th className="text-right py-3 px-4 font-medium">زمان</th>
                    <th className="text-right py-3 px-4 font-medium">نوع</th>
                    <th className="text-right py-3 px-4 font-medium">مبلغ (تومان)</th>
                    <th className="text-right py-3 px-4 font-medium">وضعیت</th>
                    <th className="text-right py-3 px-4 font-medium">توضیحات</th>
                  </tr>
                </thead>
                <tbody>
                  {transactions.map((tx) => (
                    <tr key={tx.id} className="border-b border-gray-200 dark:border-gray-800">
                      <td className="py-3 px-4 text-slate-600 dark:text-gray-400">
                        {formatTime(tx.created_at)}
                      </td>
                      <td className="py-3 px-4">
                        {tx.tx_type === 'DEPOSIT' ? 'شارژ' : 'مصرف'}
                      </td>
                      <td className="py-3 px-4 font-medium">
                        {formatCurrency(tx.amount_irr)}
                      </td>
                      <td className="py-3 px-4">
                        <span className={`inline-block px-2 py-1 text-xs rounded-full ${
                          tx.status === 'COMPLETED'
                            ? 'bg-green-500/10 text-green-500'
                            : 'bg-amber-500/10 text-amber-500'
                        }`}>
                          {tx.status === 'COMPLETED' ? 'موفق' : 'در انتظار'}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-slate-600 dark:text-gray-400">
                        {tx.description}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

