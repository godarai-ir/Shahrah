'use client';

import { Code, Book, Zap, Shield } from 'lucide-react';

export function Documentation() {
  const sections = [
    {
      icon: Zap,
      title: 'شروع سریع',
      description: 'در عرض ۵ دقیقه اولین درخواست خود را ارسال کنید',
      link: '#quick-start',
    },
    {
      icon: Code,
      title: 'مرجع API',
      description: 'مستندات کامل تمام endpoint‌ها و پارامترها',
      link: '#api-reference',
    },
    {
      icon: Book,
      title: 'راهنماها',
      description: 'آموزش‌های گام‌به‌گام برای موارد استفاده مختلف',
      link: '#guides',
    },
    {
      icon: Shield,
      title: 'امنیت',
      description: 'بهترین شیوه‌های امنیتی و مدیریت کلیدها',
      link: '#security',
    },
  ];

  return (
    <div className="py-24 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-16">
          <h1 className="text-5xl font-bold mb-6 bg-gradient-to-l from-indigo-400 to-purple-500 bg-clip-text text-transparent">
            مستندات توسعه‌دهندگان
          </h1>
          <p className="text-xl text-slate-600 dark:text-gray-400 max-w-2xl mx-auto">
            همه چیز را که برای یکپارچه‌سازی گُدار AI نیاز دارید
          </p>
        </div>

        <div className="grid md:grid-cols-2 gap-8 mb-16">
          {sections.map((section) => (
            <a
              key={section.title}
              href={section.link}
              className="p-8 rounded-2xl border border-gray-200 dark:border-gray-800 hover:border-indigo-500 dark:hover:border-indigo-500 transition-colors group"
            >
              <section.icon className="w-12 h-12 text-indigo-500 mb-4" />
              <h3 className="text-2xl font-bold mb-2 group-hover:text-indigo-500 transition-colors">
                {section.title}
              </h3>
              <p className="text-slate-600 dark:text-gray-400">{section.description}</p>
            </a>
          ))}
        </div>

        <div className="p-8 rounded-2xl border border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50">
          <h2 className="text-2xl font-bold mb-6">مثال سریع</h2>
          <pre className="bg-[#0B0F19] text-gray-300 p-6 rounded-lg overflow-x-auto" dir="ltr">
            <code>{`import OpenAI from 'openai';

const client = new OpenAI({
  apiKey: 'YOUR_SHAHRAH_API_KEY',
  baseURL: 'https://api.sharah.ir/v1'
});

const response = await client.chat.completions.create({
  model: 'gpt-4',
  messages: [
    { role: 'user', content: 'سلام! چطوری؟' }
  ]
});

console.log(response.choices[0].message.content);`}</code>
          </pre>
        </div>
      </div>
    </div>
  );
}

