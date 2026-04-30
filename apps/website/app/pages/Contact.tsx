'use client';

import { Mail, MessageSquare, Phone, MapPin } from 'lucide-react';

export function Contact() {
  return (
    <div className="py-24 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-16">
          <h1 className="text-5xl font-bold mb-6 bg-gradient-to-l from-indigo-400 to-purple-500 bg-clip-text text-transparent">
            تماس با ما
          </h1>
          <p className="text-xl text-slate-600 dark:text-gray-400 max-w-2xl mx-auto">
            سوال، پیشنهاد یا بازخوردی دارید؟ دوست داریم از شما بشنویم
          </p>
        </div>

        <div className="grid md:grid-cols-2 gap-12">
          <div>
            <h2 className="text-2xl font-bold mb-8">راه‌های ارتباطی</h2>
            <div className="space-y-6">
              {[
                {
                  icon: Mail,
                  title: 'ایمیل',
                  value: 'info@sharah.ir',
                  description: 'پاسخگویی در کمتر از ۲۴ ساعت',
                },
                {
                  icon: MessageSquare,
                  title: 'پشتیبانی آنلاین',
                  value: 'چت مستقیم',
                  description: 'دکمه پشتیبانی در پایین صفحه',
                },
                {
                  icon: Phone,
                  title: 'تلفن',
                  value: '۰۹۱۶-۶۰۹-۵۹۶۳',
                  description: 'شنبه تا چهارشنبه، ۹ صبح تا ۶ عصر',
                },
                {
                  icon: MapPin,
                  title: 'آدرس',
                  value: 'تهران، ایران',
                  description: 'کارخانه نوآوری آزادی',
                },
              ].map((contact) => (
                <div key={contact.title} className="flex gap-4 p-6 rounded-xl border border-gray-200 dark:border-gray-800">
                  <contact.icon className="w-6 h-6 text-indigo-500 shrink-0 mt-1" />
                  <div>
                    <h3 className="font-bold mb-1">{contact.title}</h3>
                    <p className="text-slate-900 dark:text-white mb-1">{contact.value}</p>
                    <p className="text-sm text-slate-600 dark:text-gray-400">{contact.description}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div>
            <h2 className="text-2xl font-bold mb-8">پیام بگذارید</h2>
            <form className="space-y-6">
              <div>
                <label className="block text-sm font-medium mb-2">نام و نام خانوادگی</label>
                <input
                  type="text"
                  className="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="نام شما"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">ایمیل</label>
                <input
                  type="email"
                  className="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="email@example.com"
                  dir="ltr"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">موضوع</label>
                <input
                  type="text"
                  className="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="موضوع پیام شما"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">پیام</label>
                <textarea
                  rows={5}
                  className="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="پیام خود را بنویسید..."
                />
              </div>
              <button
                type="submit"
                className="w-full py-3 rounded-lg bg-indigo-500 hover:bg-indigo-600 text-white font-medium transition-all shadow-[0_0_15px_rgba(99,102,241,0.3)]"
              >
                ارسال پیام
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}

