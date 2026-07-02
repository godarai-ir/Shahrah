import React from 'react';
import { motion } from 'motion/react';

type GatewayFlowProps = {
  id: string;
};

function GatewayFlow({ id }: GatewayFlowProps) {
  const requestMarkerId = `${id}-request-arrow`;
  const responseMarkerId = `${id}-response-arrow`;

  return (
    <div className="gateway-pattern__connector" aria-hidden="true">
      <svg className="gateway-pattern__flow-svg gateway-pattern__flow-svg--wide" viewBox="0 0 132 56" role="img">
        <defs>
          <marker id={requestMarkerId} markerWidth="9" markerHeight="9" refX="7.75" refY="4.5" orient="auto" markerUnits="strokeWidth">
            <path d="M 0 0 L 9 4.5 L 0 9 Z" className="gateway-pattern__arrow-head gateway-pattern__arrow-head--request" />
          </marker>
          <marker id={responseMarkerId} markerWidth="9" markerHeight="9" refX="7.75" refY="4.5" orient="auto" markerUnits="strokeWidth">
            <path d="M 0 0 L 9 4.5 L 0 9 Z" className="gateway-pattern__arrow-head gateway-pattern__arrow-head--response" />
          </marker>
        </defs>
        <path className="gateway-pattern__flow-line gateway-pattern__flow-line--request" d="M 118 18 L 14 18" markerEnd={`url(#${requestMarkerId})`} />
        <path className="gateway-pattern__flow-line gateway-pattern__flow-line--response" d="M 14 38 L 118 38" markerEnd={`url(#${responseMarkerId})`} />
        <motion.circle
          className="gateway-pattern__flow-pulse gateway-pattern__flow-pulse--request"
          r="4.5"
          cy="18"
          initial={false}
          animate={{ cx: [112, 20], opacity: [0, 1, 1, 0], scale: [0.65, 1, 1, 0.65] }}
          transition={{ duration: 1.55, repeat: Infinity, ease: 'linear' }}
        />
        <motion.circle
          className="gateway-pattern__flow-pulse gateway-pattern__flow-pulse--response"
          r="4"
          cy="38"
          initial={false}
          animate={{ cx: [20, 112], opacity: [0, 1, 1, 0], scale: [0.65, 1, 1, 0.65] }}
          transition={{ duration: 1.75, repeat: Infinity, ease: 'linear', delay: 0.12 }}
        />
      </svg>

      <svg className="gateway-pattern__flow-svg gateway-pattern__flow-svg--stacked" viewBox="0 0 96 112" role="img">
        <defs>
          <marker id={`${requestMarkerId}-stacked`} markerWidth="9" markerHeight="9" refX="7.75" refY="4.5" orient="auto" markerUnits="strokeWidth">
            <path d="M 0 0 L 9 4.5 L 0 9 Z" className="gateway-pattern__arrow-head gateway-pattern__arrow-head--request" />
          </marker>
          <marker id={`${responseMarkerId}-stacked`} markerWidth="9" markerHeight="9" refX="7.75" refY="4.5" orient="auto" markerUnits="strokeWidth">
            <path d="M 0 0 L 9 4.5 L 0 9 Z" className="gateway-pattern__arrow-head gateway-pattern__arrow-head--response" />
          </marker>
        </defs>
        <path className="gateway-pattern__flow-line gateway-pattern__flow-line--request" d="M 34 14 L 34 96" markerEnd={`url(#${requestMarkerId}-stacked)`} />
        <path className="gateway-pattern__flow-line gateway-pattern__flow-line--response" d="M 62 96 L 62 14" markerEnd={`url(#${responseMarkerId}-stacked)`} />
        <motion.circle
          className="gateway-pattern__flow-pulse gateway-pattern__flow-pulse--request"
          r="4.5"
          cx="34"
          initial={false}
          animate={{ cy: [20, 90], opacity: [0, 1, 1, 0], scale: [0.65, 1, 1, 0.65] }}
          transition={{ duration: 1.55, repeat: Infinity, ease: 'linear' }}
        />
        <motion.circle
          className="gateway-pattern__flow-pulse gateway-pattern__flow-pulse--response"
          r="4"
          cx="62"
          initial={false}
          animate={{ cy: [90, 20], opacity: [0, 1, 1, 0], scale: [0.65, 1, 1, 0.65] }}
          transition={{ duration: 1.75, repeat: Infinity, ease: 'linear', delay: 0.12 }}
        />
      </svg>
    </div>
  );
}

export default function GatewayPattern() {
  return (
    <section className="gateway-pattern" aria-label="الگوی درگاه هوش مصنوعی در GodarAI">
      <div className="gateway-pattern__zone gateway-pattern__zone--app">
        <div className="gateway-pattern__label">اپلیکیشن شما</div>
        <div className="gateway-pattern__app-icon" aria-hidden="true">
          <span>G</span>
        </div>
        <p>وب‌اپ، سرویس بک‌اند، عامل هوشمند یا ابزار داخلی</p>
      </div>

      <GatewayFlow id="gateway-app-flow" />

      <div className="gateway-pattern__zone gateway-pattern__zone--gateway">
        <div className="gateway-pattern__label">GodarAI Gateway</div>
        <div className="gateway-pattern__capabilities">
          <span>مشاهده‌پذیری</span>
          <span>مسیردهی</span>
          <span>کنترل هزینه</span>
          <span>Rate Limit</span>
          <span>Budget Limit</span>
          <span>Guardrails</span>
          <span>Virtual Model</span>
          <span>احراز هویت</span>
        </div>
      </div>

      <GatewayFlow id="gateway-provider-flow" />

      <div className="gateway-pattern__zone gateway-pattern__zone--providers">
        <div className="gateway-pattern__label">مدل‌ها و ارائه‌دهنده‌ها</div>
        <div className="gateway-pattern__provider-grid" aria-hidden="true">
          <span>LLM</span>
          <span>Embedding</span>
          <span>Vision</span>
          <span>Audio</span>
          <span>Image</span>
          <span>Realtime</span>
        </div>
        <p>چند مدل و provider پشت یک قرارداد مصرف پایدار</p>
      </div>
    </section>
  );
}
