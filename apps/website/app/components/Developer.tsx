'use client';

import { motion } from 'motion/react';
import { CircleCheck, Copy } from 'lucide-react';
import { useState } from 'react';

export function Developer() {
  const [copied, setCopied] = useState(false);
  const code = `import { OpenAI } from "openai";

const client = new OpenAI({
  apiKey: "SHAHRAH_API_KEY",
  baseURL: "https://api.shahrah.ai/v1" // تنها خطی که باید تغییر کند
});

const response = await client.chat.completions.create({
  model: "gpt-4-turbo",
  messages: [{ role: "user", content: "سلام دنیا" }]
});

console.log(response.choices[0].message.content);`;

  const handleCopy = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section id="developers" className="py-24 overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid lg:grid-cols-2 gap-16 items-center">
          
          <div>
            <h2 className="text-3xl sm:text-4xl font-bold text-slate-900 dark:text-white mb-6">
              طراحی شده برای توسعه‌دهندگان
            </h2>
            <p className="text-lg text-slate-600 dark:text-gray-400 mb-8 leading-relaxed">
              شاهراه کاملاً با استاندارد OpenAI سازگار است. هیچ نیازی به یادگیری SDK جدید یا تغییرات گسترده در کدهای فعلی‌تان نیست. فقط Base URL را تغییر دهید.
            </p>
            
            <ul className="space-y-4 mb-8">
              {[
                'مهاجرت (Migration) در کمتر از ۱ دقیقه',
                'سازگاری ۱۰۰٪ با کتابخانه‌های رسمی OpenAI',
                'پشتیبانی از Streaming و Function Calling',
                'مستندات کامل با مثال‌های واقعی'
              ].map((item, i) => (
                <li key={i} className="flex items-center gap-3 text-slate-700 dark:text-gray-300">
                  <CircleCheck className="w-5 h-5 text-green-500 flex-shrink-0" />
                  <span>{item}</span>
                </li>
              ))}
            </ul>
            
            <a href="#docs" className="inline-flex items-center justify-center px-6 py-3 rounded-xl bg-slate-900 dark:bg-white text-white dark:text-slate-900 font-medium hover:bg-slate-800 dark:hover:bg-gray-100 transition-colors">
              مطالعه مستندات API
            </a>
          </div>

          <div className="relative" dir="ltr">
            <div className="absolute inset-0 bg-gradient-to-tr from-indigo-500 to-purple-500 rounded-2xl blur-2xl opacity-20"></div>
            <div className="relative bg-[#0B0F19] rounded-2xl border border-gray-800 shadow-2xl overflow-hidden">
              <div className="flex items-center justify-between px-4 py-3 border-b border-gray-800 bg-[#111827]">
                <div className="flex space-x-2">
                  <div className="w-3 h-3 rounded-full bg-gray-700"></div>
                  <div className="w-3 h-3 rounded-full bg-gray-700"></div>
                  <div className="w-3 h-3 rounded-full bg-gray-700"></div>
                </div>
                <div className="text-xs text-gray-400 font-mono">index.ts</div>
                <button 
                  onClick={handleCopy}
                  className="text-gray-400 hover:text-white transition-colors"
                  title="Copy code"
                >
                  <Copy className="w-4 h-4" />
                </button>
              </div>
              <div className="p-6 text-sm font-mono text-gray-300 overflow-x-auto relative">
                {copied && (
                  <div className="absolute top-4 right-4 bg-green-500/20 text-green-400 px-3 py-1 rounded text-xs">
                    Copied!
                  </div>
                )}
                <pre>
                  <code>
<span className="text-purple-400">import</span> {'{'} OpenAI {'}'} <span className="text-purple-400">from</span> <span className="text-green-400">"openai"</span>;{'\n\n'}
<span className="text-purple-400">const</span> client = <span className="text-purple-400">new</span> OpenAI({'{'}{'\n'}
{'  '}apiKey: <span className="text-green-400">"SHAHRAH_API_KEY"</span>,{'\n'}
{'  '}baseURL: <span className="text-green-400">"https://api.shahrah.ai/v1"</span> <span className="text-gray-500">// تنها خطی که باید تغییر کند</span>{'\n'}
{'}'});{'\n\n'}
<span className="text-purple-400">const</span> response = <span className="text-blue-400">await</span> client.chat.completions.create({'{'}{'\n'}
{'  '}model: <span className="text-green-400">"gpt-4-turbo"</span>,{'\n'}
{'  '}messages: [{'{'} role: <span className="text-green-400">"user"</span>, content: <span className="text-green-400">"سلام دنیا"</span> {'}'}]{'\n'}
{'}'});{'\n\n'}
<span className="text-yellow-400">console</span>.<span className="text-blue-400">log</span>(response.choices[<span className="text-orange-400">0</span>].message.content);
                  </code>
                </pre>
              </div>
            </div>
          </div>

        </div>
      </div>
    </section>
  );
}

