'use client';

export function Privacy() {
  return (
    <div className="py-24 px-4 sm:px-6 lg:px-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold mb-8">سیاست حفظ حریم خصوصی</h1>
        <div className="prose prose-slate dark:prose-invert max-w-none space-y-8">
          <section>
            <h2 className="text-2xl font-bold mb-4">۱. جمع‌آوری اطلاعات</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              ما اطلاعات شخصی شما را تنها در صورت نیاز و با رضایت شما جمع‌آوری می‌کنیم. این اطلاعات شامل
              نام، ایمیل، و اطلاعات پرداخت می‌شود که برای ارائه خدمات به شما ضروری است.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۲. استفاده از اطلاعات</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed mb-4">
              اطلاعات جمع‌آوری شده برای اهداف زیر استفاده می‌شود:
            </p>
            <ul className="list-disc list-inside space-y-2 text-slate-600 dark:text-gray-400 mr-4">
              <li>ارائه و بهبود خدمات</li>
              <li>پردازش پرداخت‌ها و صدور صورتحساب</li>
              <li>ارسال اطلاعیه‌های مهم سرویس</li>
              <li>پشتیبانی از کاربران</li>
              <li>تحلیل و بهینه‌سازی عملکرد پلتفرم</li>
            </ul>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۳. محافظت از داده‌ها</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              ما از روش‌های امنیتی پیشرفته برای محافظت از اطلاعات شما استفاده می‌کنیم، از جمله رمزنگاری،
              احراز هویت چند عاملی، و نظارت مستمر بر امنیت سیستم‌ها.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۴. اشتراک‌گذاری اطلاعات</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              ما اطلاعات شخصی شما را با اشخاص ثالث به اشتراک نمی‌گذاریم، مگر در موارد زیر:
            </p>
            <ul className="list-disc list-inside space-y-2 text-slate-600 dark:text-gray-400 mr-4 mt-4">
              <li>با رضایت صریح شما</li>
              <li>برای رعایت قوانین و مقررات</li>
              <li>با ارائه‌دهندگان خدمات معتمد که قراردادهای محرمانگی دارند</li>
            </ul>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۵. حقوق کاربران</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              شما حق دارید به اطلاعات خود دسترسی داشته باشید، آن‌ها را ویرایش کنید، یا حساب خود را حذف کنید.
              برای استفاده از این حقوق، با تیم پشتیبانی ما تماس بگیرید.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۶. کوکی‌ها</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              ما از کوکی‌ها برای بهبود تجربه کاربری، تحلیل ترافیک، و شخصی‌سازی محتوا استفاده می‌کنیم.
              شما می‌توانید کوکی‌ها را از طریق تنظیمات مرورگر خود مدیریت کنید.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۷. تغییرات در سیاست</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              ما ممکن است این سیاست را به‌روزرسانی کنیم. تغییرات مهم از طریق ایمیل یا اطلاعیه در پلتفرم
              به اطلاع شما خواهد رسید.
            </p>
          </section>

          <section>
            <h2 className="text-2xl font-bold mb-4">۸. تماس با ما</h2>
            <p className="text-slate-600 dark:text-gray-400 leading-relaxed">
              برای سوالات مربوط به حریم خصوصی، با ما از طریق ایمیل{' '}
              <a href="mailto:privacy@sharah.ir" className="text-indigo-500 hover:text-indigo-600">
                privacy@sharah.ir
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

