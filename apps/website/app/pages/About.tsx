'use client';

import { Target, Users, Lightbulb, Heart } from 'lucide-react';

export function About() {
  return (
    <div className="py-24 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-16">
          <h1 className="text-5xl font-bold mb-6 bg-gradient-to-l from-indigo-400 to-purple-500 bg-clip-text text-transparent">
            درباره گدارAI
          </h1>
          <p className="text-xl text-slate-600 dark:text-gray-400 max-w-2xl mx-auto">
            ما مسیر اتصال تیم‌های ایرانی به مدل‌های زبانی را ساده‌تر، قابل‌پیگیری‌تر و قابل‌کنترل‌تر می‌کنیم.
          </p>
        </div>

        <div className="grid md:grid-cols-2 gap-12 mb-20">
          <div>
            <h2 className="text-3xl font-bold mb-6">ماموریت ما</h2>
            <p className="text-lg text-slate-600 dark:text-gray-400 leading-relaxed mb-4">
              گدارAI با هدف حل یکی از چالش‌های جدی تیم‌های فنی ایرانی شکل گرفت:
              اتصال پایدارتر به مدل‌های زبانی، پرداخت ریالی، کنترل هزینه و پایش مصرف در یک مسیر واحد.
            </p>
            <p className="text-lg text-slate-600 dark:text-gray-400 leading-relaxed">
              ما تلاش می‌کنیم محدودیت‌های پرداخت، پراکندگی کلیدها و نبود مشاهده‌پذیری مانع ساخت محصول‌های مبتنی بر هوش مصنوعی نشود.
            </p>
          </div>
          <div>
            <h2 className="text-3xl font-bold mb-6">چشم‌انداز ما</h2>
            <p className="text-lg text-slate-600 dark:text-gray-400 leading-relaxed mb-4">
              می‌خواهیم هر تیم محصول بتواند مدل مناسب خود را انتخاب کند، مصرف را ببیند، هزینه را کنترل کند و درباره مسیر هر درخواست پاسخ روشن داشته باشد.
            </p>
            <p className="text-lg text-slate-600 dark:text-gray-400 leading-relaxed">
              گدارAI قرار است یک لایه عملیاتی قابل اتکا برای استفاده از مدل‌های زبانی در بازار ایران باشد.
            </p>
          </div>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
          {[
            {
              icon: Target,
              title: 'هدف‌محور',
              description: 'تمرکز ما روی حل مشکلات واقعی توسعه‌دهندگان است',
            },
            {
              icon: Users,
              title: 'کاربرمحور',
              description: 'تجربه کاربری ساده و روان در اولویت ما قرار دارد',
            },
            {
              icon: Lightbulb,
              title: 'دقیق',
              description: 'ادعاها، هزینه‌ها و محدودیت‌ها را شفاف بیان می‌کنیم',
            },
            {
              icon: Heart,
              title: 'مسئولیت‌پذیر',
              description: 'متعهد به ارائه سرویسی پایدار، قابل بررسی و قابل پشتیبانی',
            },
          ].map((value) => (
            <div key={value.title} className="text-center p-6">
              <value.icon className="w-12 h-12 text-indigo-500 mx-auto mb-4" />
              <h3 className="text-xl font-bold mb-2">{value.title}</h3>
              <p className="text-slate-600 dark:text-gray-400">{value.description}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

