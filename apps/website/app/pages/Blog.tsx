'use client';

import { Calendar, Clock, ArrowLeft } from 'lucide-react';

export function Blog() {
  const posts = [
    {
      title: 'چگونه گُدار AI هزینه‌های AI شما را تا ۷۰٪ کاهش می‌دهد',
      excerpt: 'با استفاده از مسیریابی هوشمند و بهینه‌سازی خودکار، می‌توانید هزینه‌های استفاده از مدل‌های هوش مصنوعی را به‌طور چشمگیری کاهش دهید.',
      date: '۱۴۰۳/۰۱/۱۵',
      readTime: '۵ دقیقه',
      category: 'بهینه‌سازی',
    },
    {
      title: 'راهنمای کامل انتخاب بهترین مدل AI برای پروژه شما',
      excerpt: 'مدل‌های مختلف هوش مصنوعی برای کاربردهای مختلف طراحی شده‌اند. این راهنما به شما کمک می‌کند بهترین انتخاب را داشته باشید.',
      date: '۱۴۰۳/۰۱/۱۰',
      readTime: '۸ دقیقه',
      category: 'آموزش',
    },
    {
      title: 'امنیت در استفاده از API‌های هوش مصنوعی',
      excerpt: 'بهترین شیوه‌های امنیتی برای محافظت از کلیدهای API، داده‌های کاربران و جلوگیری از حملات رایج.',
      date: '۱۴۰۳/۰۱/۰۵',
      readTime: '۶ دقیقه',
      category: 'امنیت',
    },
    {
      title: 'معرفی قابلیت مسیریابی هوشمند در گُدار AI',
      excerpt: 'با مسیریابی هوشمند، درخواست‌های شما به‌طور خودکار به بهترین و مقرون‌به‌صرفه‌ترین مدل هدایت می‌شوند.',
      date: '۱۴۰۲/۱۲/۲۸',
      readTime: '۴ دقیقه',
      category: 'محصول',
    },
    {
      title: 'چالش‌های توسعه‌دهندگان ایرانی در دسترسی به AI',
      excerpt: 'نگاهی به مشکلات و محدودیت‌هایی که توسعه‌دهندگان ایرانی برای استفاده از خدمات هوش مصنوعی با آن روبرو هستند.',
      date: '۱۴۰۲/۱۲/۲۰',
      readTime: '۷ دقیقه',
      category: 'بحث و گفتگو',
    },
    {
      title: 'آینده هوش مصنوعی در ایران',
      excerpt: 'پیش‌بینی روندها و فرصت‌های جدید در حوزه هوش مصنوعی برای کسب‌وکارها و استارتاپ‌های ایرانی.',
      date: '۱۴۰۲/۱۲/۱۵',
      readTime: '۱۰ دقیقه',
      category: 'بحث و گفتگو',
    },
  ];

  return (
    <div className="py-24 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-16">
          <h1 className="text-5xl font-bold mb-6 bg-gradient-to-l from-indigo-400 to-purple-500 bg-clip-text text-transparent">
            وبلاگ گُدار AI
          </h1>
          <p className="text-xl text-slate-600 dark:text-gray-400 max-w-2xl mx-auto">
            آخرین اخبار، آموزش‌ها و بینش‌ها در حوزه هوش مصنوعی
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
          {posts.map((post) => (
            <article
              key={post.title}
              className="group p-6 rounded-2xl border border-gray-200 dark:border-gray-800 hover:border-indigo-500 dark:hover:border-indigo-500 transition-colors cursor-pointer"
            >
              <div className="mb-4">
                <span className="inline-block px-3 py-1 text-xs font-medium rounded-full bg-indigo-500/10 text-indigo-500">
                  {post.category}
                </span>
              </div>
              <h2 className="text-xl font-bold mb-3 group-hover:text-indigo-500 transition-colors">
                {post.title}
              </h2>
              <p className="text-slate-600 dark:text-gray-400 mb-4 line-clamp-3">
                {post.excerpt}
              </p>
              <div className="flex items-center justify-between text-sm text-slate-600 dark:text-gray-400 mb-4">
                <div className="flex items-center gap-2">
                  <Calendar className="w-4 h-4" />
                  <span>{post.date}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Clock className="w-4 h-4" />
                  <span>{post.readTime}</span>
                </div>
              </div>
              <div className="flex items-center gap-2 text-indigo-500 font-medium group-hover:gap-3 transition-all">
                <span>ادامه مطلب</span>
                <ArrowLeft className="w-4 h-4" />
              </div>
            </article>
          ))}
        </div>
      </div>
    </div>
  );
}

