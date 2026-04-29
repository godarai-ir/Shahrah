'use client';

import { useState } from 'react';
import { ChevronDown, ChevronUp } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';

interface FAQItemProps {
  question: string;
  answer: string;
}

interface FAQAccordionProps {
  items: FAQItemProps[];
  className?: string;
}

export function FAQAccordion({ items, className = '' }: FAQAccordionProps) {
  const [openIndex, setOpenIndex] = useState<number | null>(null);

  return (
    <div className={`space-y-4 ${className}`}>
      {items.map((faq, idx) => {
        const isOpen = openIndex === idx;
        
        return (
          <div key={idx} className="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden bg-white dark:bg-slate-900">
            <button 
              className="w-full px-6 py-4 flex items-center justify-between bg-slate-50 hover:bg-slate-100 dark:bg-slate-900 dark:hover:bg-slate-800/80 transition-colors text-right"
              onClick={() => setOpenIndex(isOpen ? null : idx)}
            >
              <span className="font-bold text-slate-900 dark:text-white">{faq.question}</span>
              {isOpen ? <ChevronUp className="w-5 h-5 text-slate-500 shrink-0" /> : <ChevronDown className="w-5 h-5 text-slate-500 shrink-0" />}
            </button>
            <AnimatePresence>
              {isOpen && (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: 'auto', opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  transition={{ duration: 0.3, ease: 'easeInOut' }}
                  className="overflow-hidden"
                >
                  <div className="px-6 py-4 border-t border-slate-200 dark:border-slate-800 text-slate-600 dark:text-slate-400 text-sm leading-relaxed">
                    {faq.answer}
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        );
      })}
    </div>
  );
}
