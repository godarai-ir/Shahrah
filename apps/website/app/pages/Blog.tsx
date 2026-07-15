'use client';

import { Calendar, Clock, ArrowLeft } from 'lucide-react';

export function Blog() {
  const posts = [
    {
      title: 'چطور هزینه مصرف مدل‌های زبانی را قابل پیش‌بینی کنیم؟',
      excerpt: 'با مدل مجازی، سقف هزینه، حافظه نهان و گزارش مصرف می‌توانید هزینه هر تیم یا سرویس را دقیق‌تر پیگیری کنید.',
      date: '۱۴۰۳/۰۱/۱۵',
      readTime: '۵ دقیقه',
      category: 'بهینه‌سازی',
    },
    {
      title: 'چطور مدل مناسب سناریوی محصول را انتخاب کنیم؟',
      excerpt: 'برای پشتیبانی مشتری، تولید محتوا یا تحلیل پیام‌های فارسی، معیارهایی مثل کیفیت پاسخ، هزینه و مهلت پاسخ را کنار هم ببینید.',
      date: '۱۴۰۳/۰۱/۱۰',
      readTime: '۸ دقیقه',
      category: 'آموزش',
    },
    {
      title: 'امنیت در استفاده از API‌های هوش مصنوعی',
      excerpt: 'روش‌های عملی برای محافظت از توکن‌های دسترسی، داده‌های کاربران و مسیر درخواست‌های مدل.',
      date: '۱۴۰۳/۰۱/۰۵',
      readTime: '۶ دقیقه',
      category: 'امنیت',
    },
    {
      title: 'مسیریابی در گدارAI چطور کار می‌کند؟',
      excerpt: 'با مدل مجازی، اولویت، وزن و مسیر جایگزین می‌توانید رفتار درخواست‌ها را بدون تغییر مداوم کد کنترل کنید.',
      date: '۱۴۰۲/۱۲/۲۸',
      readTime: '۴ دقیقه',
      category: 'محصول',
    },
    {
      title: 'چالش‌های تیم‌های ایرانی در اتصال به مدل‌های زبانی',
      excerpt: 'پرداخت، کلیدهای متعدد، خطای ارائه‌دهنده و نبود گزارش مصرف از مسئله‌های رایج تیم‌های فنی است.',
      date: '۱۴۰۲/۱۲/۲۰',
      readTime: '۷ دقیقه',
      category: 'بحث و گفتگو',
    },
    {
      title: 'از آزمایش تا محیط عملیاتی؛ مسیر بلوغ محصول‌های هوش مصنوعی',
      excerpt: 'وقتی یک قابلیت از نمونه اولیه به محصول واقعی می‌رسد، کنترل هزینه، گاردریل و مشاهده‌پذیری ضروری‌تر می‌شود.',
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
            وبلاگ گدارAI
          </h1>
          <p className="text-xl text-slate-600 dark:text-gray-400 max-w-2xl mx-auto">
            یادداشت‌های فنی درباره درگاه مدل‌های زبانی، هزینه، امنیت و تجربه ساخت محصول در ایران
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

