'use client';

import { Target, Users, Lightbulb, Heart } from 'lucide-react';

export function About() {
  return (
    <div className="py-24 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-16">
          <h1 className="text-5xl font-bold mb-6 bg-gradient-to-l from-indigo-400 to-purple-500 bg-clip-text text-transparent">
            درباره شاهراه
          </h1>
          <p className="text-xl text-slate-600 dark:text-gray-400 max-w-2xl mx-auto">
            ما در حال ساختن پلی میان توسعه‌دهندگان ایرانی و دنیای هوش مصنوعی هستیم
          </p>
        </div>

        <div className="grid md:grid-cols-2 gap-12 mb-20">
          <div>
            <h2 className="text-3xl font-bold mb-6">ماموریت ما</h2>
            <p className="text-lg text-slate-600 dark:text-gray-400 leading-relaxed mb-4">
              شاهراه با هدف حل یکی از بزرگترین چالش‌های توسعه‌دهندگان و کسب‌وکارهای ایرانی ایجاد شد:
              دسترسی آسان، مقرون‌به‌صرفه و بدون محدودیت به پیشرفته‌ترین مدل‌های هوش مصنوعی.
            </p>
            <p className="text-lg text-slate-600 dark:text-gray-400 leading-relaxed">
              ما معتقدیم که تحریم‌ها و محدودیت‌های پرداخت نباید مانع از نوآوری و پیشرفت باشند.
            </p>
          </div>
          <div>
            <h2 className="text-3xl font-bold mb-6">چشم‌انداز ما</h2>
            <p className="text-lg text-slate-600 dark:text-gray-400 leading-relaxed mb-4">
              ما تصویر می‌کنیم که هر توسعه‌دهنده و کسب‌وکار ایرانی بتواند با اطمینان و بدون دغدغه،
              از قدرت هوش مصنوعی برای ساخت محصولات و خدمات نوآورانه استفاده کند.
            </p>
            <p className="text-lg text-slate-600 dark:text-gray-400 leading-relaxed">
              شاهراه قرار است به زیرساخت اصلی دسترسی به AI در ایران تبدیل شود.
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
              title: 'نوآور',
              description: 'همیشه به دنبال راه‌های بهتر و کارآمدتر هستیم',
            },
            {
              icon: Heart,
              title: 'مسئولیت‌پذیر',
              description: 'متعهد به ارائه سرویسی پایدار و قابل اعتماد',
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

