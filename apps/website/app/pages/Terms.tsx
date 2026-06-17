'use client';

export function Terms() {
  return (
    <div className="py-24 px-4 sm:px-6 lg:px-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold mb-8">شرایط و قوانین استفاده</h1>
        <div className="prose prose-slate dark:prose-invert max-w-none space-y-8">
          <section>
            <h2 className="text-2xl font-bold mb-4">۱. پذیرش شرایط</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              با استفاده از خدمات گُدار AI، شما این شرایط و قوانین را می‌پذیرید. اگر با این شرایط موافق نیستید،
              لطفاً از خدمات استفاده نکنید.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۲. تعریف خدمات</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              گُدار AI یک پلتفرم دروازه API است که دسترسی یکپارچه به مدل‌های مختلف هوش مصنوعی را فراهم می‌کند.
              خدمات شامل API Gateway، مسیریابی هوشمند، بهینه‌سازی هزینه، و داشبورد مدیریت است.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۳. حساب کاربری</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed mb-4">
              برای استفاده از خدمات، باید یک حساب کاربری ایجاد کنید. شما مسئول هستید:
            </p>
            <ul className="list-disc list-inside space-y-2 text-slate-600 dark:text-gray-400 mr-4">
              <li>حفظ امنیت رمز عبور و کلیدهای API خود</li>
              <li>ارائه اطلاعات صحیح و به‌روز</li>
              <li>اطلاع‌رسانی فوری در صورت استفاده غیرمجاز از حساب</li>
              <li>تمامی فعالیت‌هایی که تحت حساب شما انجام می‌شود</li>
            </ul>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۴. استفاده مجاز</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed mb-4">
              شما موظف به استفاده قانونی و اخلاقی از خدمات هستید. موارد زیر ممنوع است:
            </p>
            <ul className="list-disc list-inside space-y-2 text-slate-600 dark:text-gray-400 mr-4">
              <li>نقض قوانین ایران یا بین‌المللی</li>
              <li>تولید محتوای مضر، تهدیدآمیز، یا غیرقانونی</li>
              <li>سوءاستفاده از سیستم یا تلاش برای دور زدن محدودیت‌ها</li>
              <li>اشتراک‌گذاری کلیدهای API با اشخاص غیرمجاز</li>
              <li>استفاده برای spam یا فعالیت‌های مخرب</li>
            </ul>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۵. پرداخت و قیمت‌گذاری</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              هزینه‌های خدمات بر اساس پلن انتخابی و میزان استفاده محاسبه می‌شود. پرداخت‌ها به تومان و از
              طریق درگاه‌های پرداخت معتبر ایرانی انجام می‌شود. ما حق تغییر قیمت‌ها را با اطلاع‌رسانی ۳۰ روزه
              محفوظ می‌داریم.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۶. محدودیت‌ها و سطح سرویس</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              هر پلن دارای محدودیت‌های مشخصی برای تعداد درخواست، توکن، و سایر منابع است. ما تلاش می‌کنیم
              uptime بالایی را حفظ کنیم، اما نمی‌توانیم دسترسی ۱۰۰٪ را تضمین کنیم مگر در پلن‌های سازمانی با SLA.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۷. مالکیت معنوی</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              تمامی حقوق مالکیت معنوی پلتفرم متعلق به گُدار AI است. محتوای تولید شده توسط شما از طریق API
              متعلق به شماست، اما ما حق ذخیره و استفاده از داده‌های ناشناس برای بهبود خدمات را داریم.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۸. محدودیت مسئولیت</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              گُدار AI در قبال خسارات غیرمستقیم، از دست دادن داده، یا اختلال در کسب‌وکار مسئول نیست. مسئولیت
              ما محدود به مبلغ پرداختی شما در ۱۲ ماه گذشته است.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۹. تعلیق و خاتمه خدمات</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              ما حق تعلیق یا خاتمه حساب شما را در صورت نقض این شرایط، عدم پرداخت، یا استفاده غیرمجاز داریم.
              شما می‌توانید در هر زمان حساب خود را لغو کنید.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۱۰. تغییرات در شرایط</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              ما ممکن است این شرایط را به‌روزرسانی کنیم. ادامه استفاده از خدمات پس از تغییرات به معنای
              پذیرش شرایط جدید است.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۱۱. قانون حاکم</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              این قرارداد تحت قوانین جمهوری اسلامی ایران است و اختلافات در دادگاه‌های صالح تهران حل و فصل
              می‌شود.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۱۲. تماس با ما</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              برای سوالات مربوط به شرایط استفاده، با ما از طریق ایمیل{' '}
              <a href="mailto:legal@sharah.ir" className="text-indigo-500 hover:text-indigo-600">
                legal@sharah.ir
              </a>{' '}
              تماس بگیرید.
            </p>
          </section>

          <p className="text-sm text-slate-600 dark:text-gray-400 mt-12">
            آخرین به‌روزرسانی: ۱۴۰۳/۰۱/۲۲
          </p>
        </div>
      </div>
    </div>
  );
}

