'use client';

import { Check } from 'lucide-react';

export function Pricing() {
  const plans = [
    {
      name: 'رایگان',
      price: '۰',
      description: 'برای شروع و آزمایش',
      features: [
        '۱۰۰،۰۰۰ توکن ماهانه',
        'دسترسی به ۳ مدل هوش مصنوعی',
        'پشتیبانی ایمیلی',
        'داشبورد پایه',
      ],
    },
    {
      name: 'حرفه‌ای',
      price: '۲۹۹،۰۰۰',
      description: 'برای تیم‌های کوچک',
      features: [
        '۵ میلیون توکن ماهانه',
        'دسترسی به تمام مدل‌ها',
        'مسیریابی هوشمند',
        'بهینه‌سازی هزینه',
        'پشتیبانی اولویت‌دار',
        'تحلیل‌های پیشرفته',
      ],
      popular: true,
    },
    {
      name: 'سازمانی',
      price: 'تماس بگیرید',
      description: 'برای شرکت‌های بزرگ',
      features: [
        'توکن نامحدود',
        'مدل‌های اختصاصی',
        'SLA تضمین‌شده',
        'مدیریت تیم',
        'پشتیبانی ۲۴/۷',
        'مشاوره اختصاصی',
        'امنیت سازمانی',
      ],
    },
  ];

  return (
    <div className="py-24 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-16">
          <h1 className="text-5xl font-bold mb-6 bg-gradient-to-l from-indigo-400 to-purple-500 bg-clip-text text-transparent">
            قیمت‌گذاری شفاف و منصفانه
          </h1>
          <p className="text-xl text-slate-600 dark:text-gray-400 max-w-2xl mx-auto">
            با پرداخت به تومان، بدون دغدغه تحریم و محدودیت
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-8">
          {plans.map((plan) => (
            <div
              key={plan.name}
              className={`relative p-8 rounded-2xl border ${
                plan.popular
                  ? 'border-indigo-500 bg-indigo-500/5 dark:bg-indigo-500/10'
                  : 'border-gray-200 dark:border-gray-800'
              }`}
            >
              {plan.popular && (
                <div className="absolute -top-4 right-1/2 translate-x-1/2 px-4 py-1 bg-indigo-500 text-white text-sm font-medium rounded-full">
                  محبوب‌ترین
                </div>
              )}
              <div className="mb-8">
                <h3 className="text-2xl font-bold mb-2">{plan.name}</h3>
                <p className="text-slate-600 dark:text-gray-400 mb-4">{plan.description}</p>
                <div className="flex items-baseline gap-2">
                  <span className="text-4xl font-bold">{plan.price}</span>
                  {plan.price !== 'تماس بگیرید' && <span className="text-slate-600 dark:text-gray-400">تومان/ماه</span>}
                </div>
              </div>
              <ul className="space-y-4 mb-8">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-start gap-3">
                    <Check className="w-5 h-5 text-indigo-500 shrink-0 mt-0.5" />
                    <span className="text-slate-600 dark:text-gray-300">{feature}</span>
                  </li>
                ))}
              </ul>
              <button
                className={`w-full py-3 rounded-lg font-medium transition-all ${
                  plan.popular
                    ? 'bg-indigo-500 hover:bg-indigo-600 text-white shadow-[0_0_15px_rgba(99,102,241,0.3)]'
                    : 'bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700'
                }`}
              >
                {plan.price === 'تماس بگیرید' ? 'تماس با فروش' : 'شروع کنید'}
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

