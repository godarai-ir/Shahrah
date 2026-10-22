'use client';

import { ChevronDown, Menu, Moon, Sun, X } from 'lucide-react';
import Link from 'next/link';
import { useEffect, useRef, useState } from 'react';
import type { ReactNode } from 'react';
import { navMenus, pricingLink, visibleColumns, visibleLinks } from '@/app/lib/navigation';
import type { NavLink, NavMenu } from '@/app/lib/navigation';
import { toneSurface } from '@/app/lib/pillars';
import { GodarAILogo } from './GodarAILogo';
import { useTheme } from './ThemeProvider';

const focusRing = 'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-accent';

const panelWidth = ['', 'w-80 grid-cols-1', 'w-160 grid-cols-2', 'w-200 grid-cols-3'];

function NavAnchor({ item, className, onNavigate, children }: { item: NavLink; className: string; onNavigate?: () => void; children: ReactNode }) {
  if (item.external) {
    return (
      <a href={item.href} className={className} onClick={onNavigate}>
        {children}
      </a>
    );
  }

  return (
    <Link href={item.href} className={className} onClick={onNavigate}>
      {children}
    </Link>
  );
}

function MenuRow({ item, onNavigate }: { item: NavLink; onNavigate?: () => void }) {
  const well = item.pillar ? toneSurface[item.pillar] : 'bg-site-panel-deep';

  return (
    <NavAnchor
      item={item}
      onNavigate={onNavigate}
      className={`flex min-h-12 items-start gap-3 rounded-2xl p-3 transition-colors hover:bg-site-panel ${focusRing}`}
    >
      {item.icon ? (
        <span className={`flex size-9 shrink-0 items-center justify-center rounded-xl text-site-accent ${well}`}>
          <item.icon className="size-4" />
        </span>
      ) : null}
      <span className="flex min-w-0 flex-col gap-0.5">
        <span className="text-sm font-semibold text-site-ink">{item.label}</span>
        {item.description ? <span className="text-xs leading-6 text-site-ink-soft">{item.description}</span> : null}
      </span>
    </NavAnchor>
  );
}

function MegaPanel({ menu, onNavigate }: { menu: NavMenu; onNavigate: () => void }) {
  const columns = visibleColumns(menu.columns);

  return (
    <div className={`grid gap-4 rounded-3xl border border-site-line bg-site-ground p-4 shadow-card ${panelWidth[Math.min(columns.length, 3)]}`}>
      {columns.map((column, index) => (
        <div key={column.title ?? index} className={`flex flex-col gap-1 ${index > 0 ? 'border-s border-site-line ps-4' : ''}`}>
          {column.title ? <p className="m-0 px-3 pb-1 pt-2 text-xs font-medium text-site-ink-soft">{column.title}</p> : null}
          {column.links.map((item) => (
            <MenuRow key={item.href} item={item} onNavigate={onNavigate} />
          ))}
        </div>
      ))}
    </div>
  );
}

function ThemeToggle() {
  const { resolvedTheme, setTheme } = useTheme();
  const dark = resolvedTheme === 'dark';

  return (
    <button
      type="button"
      aria-label={dark ? 'تغییر به حالت روشن' : 'تغییر به حالت تاریک'}
      onClick={() => setTheme(dark ? 'light' : 'dark')}
      className={`flex size-11 items-center justify-center rounded-full text-site-ink-soft transition-colors hover:bg-site-panel hover:text-site-ink ${focusRing}`}
    >
      {dark ? <Sun className="size-5" /> : <Moon className="size-5" />}
    </button>
  );
}

export function Navbar() {
  const [openMenu, setOpenMenu] = useState<string | null>(null);
  const [sheetOpen, setSheetOpen] = useState(false);
  const navRef = useRef<HTMLElement>(null);

  const closeMenu = (returnFocus = false) => {
    if (returnFocus && openMenu) {
      document.getElementById(`nav-button-${openMenu}`)?.focus();
    }

    setOpenMenu(null);
  };

  // Escape and a click outside close the desktop menu; the mobile sheet also locks page scroll while open.
  useEffect(() => {
    if (!openMenu && !sheetOpen) {
      return;
    }

    const onKey = (event: KeyboardEvent) => {
      if (event.key !== 'Escape') {
        return;
      }

      if (openMenu) {
        document.getElementById(`nav-button-${openMenu}`)?.focus();
        setOpenMenu(null);
      }

      if (sheetOpen) {
        setSheetOpen(false);
        document.getElementById('nav-sheet-button')?.focus();
      }
    };

    const onPointer = (event: PointerEvent) => {
      if (openMenu && !navRef.current?.contains(event.target as Node)) {
        setOpenMenu(null);
      }
    };

    document.addEventListener('keydown', onKey);
    document.addEventListener('pointerdown', onPointer);

    return () => {
      document.removeEventListener('keydown', onKey);
      document.removeEventListener('pointerdown', onPointer);
    };
  }, [openMenu, sheetOpen]);

  useEffect(() => {
    if (!sheetOpen) {
      return;
    }

    const previous = document.body.style.overflow;
    const desktop = window.matchMedia('(min-width: 1024px)');
    const onChange = () => desktop.matches && setSheetOpen(false);

    document.body.style.overflow = 'hidden';
    desktop.addEventListener('change', onChange);

    return () => {
      document.body.style.overflow = previous;
      desktop.removeEventListener('change', onChange);
    };
  }, [sheetOpen]);

  const closeSheet = () => setSheetOpen(false);

  return (
    <nav ref={navRef} aria-label="ناوبری اصلی" className="sticky top-0 z-50 w-full border-b border-site-line bg-site-ground">
      <div className="mx-auto flex min-h-16 max-w-7xl items-center justify-between gap-3 px-4 sm:px-6 lg:px-8">
        <div className="flex min-w-0 items-center gap-6 xl:gap-10">
          <Link href="/" className={`flex min-w-0 items-center gap-2 rounded-xl ${focusRing}`}>
            <GodarAILogo decorative priority className="h-9 w-auto max-w-12 shrink-0 object-contain" sizes="3rem" />
            <span className="truncate text-lg font-bold text-site-ink sm:text-xl">گدارAI</span>
          </Link>

          {/* Panels anchor to the list, not their button, so even the widest one stays inside a 1024px viewport. */}
          <ul className="relative m-0 hidden list-none items-center gap-1 p-0 lg:flex">
            {navMenus.map((menu) => {
              const open = openMenu === menu.id;

              return (
                <li
                  key={menu.id}
                  onMouseEnter={() => setOpenMenu(menu.id)}
                  onMouseLeave={() => closeMenu()}
                >
                  <button
                    type="button"
                    id={`nav-button-${menu.id}`}
                    aria-expanded={open}
                    aria-controls={`nav-panel-${menu.id}`}
                    // Hover already opened the menu, so a mouse click (detail > 0) keeps it open; Enter or Space (detail 0) toggles.
                    onClick={(event) => setOpenMenu(open && event.detail === 0 ? null : menu.id)}
                    className={`flex h-10 items-center gap-1 rounded-full px-3 text-sm font-medium transition-colors hover:text-site-ink ${
                      open ? 'text-site-ink' : 'text-site-ink-soft'
                    } ${focusRing}`}
                  >
                    {menu.label}
                    <ChevronDown aria-hidden="true" className={`size-4 transition-transform duration-300 ${open ? 'rotate-180' : ''}`} />
                  </button>
                  <div id={`nav-panel-${menu.id}`} hidden={!open} className="absolute start-0 top-full pt-2">
                    <MegaPanel menu={menu} onNavigate={() => setOpenMenu(null)} />
                  </div>
                </li>
              );
            })}
            <li>
              <Link
                href={pricingLink.href}
                className={`flex h-10 items-center rounded-full px-3 text-sm font-medium text-site-ink-soft transition-colors hover:text-site-ink ${focusRing}`}
              >
                {pricingLink.label}
              </Link>
            </li>
          </ul>
        </div>

        <div className="hidden items-center gap-1 lg:flex">
          <ThemeToggle />
          <Link href="/login" className={`flex h-10 items-center rounded-full px-3 text-sm font-medium text-site-ink-soft transition-colors hover:text-site-ink ${focusRing}`}>
            ورود
          </Link>
          <Link href="/demo" className={`flex h-10 items-center rounded-full px-3 text-sm font-medium text-site-ink-soft transition-colors hover:text-site-ink ${focusRing}`}>
            درخواست دمو
          </Link>
          <Link
            href="/onboarding"
            className={`ms-2 flex h-10 items-center rounded-full bg-site-accent px-5 text-sm font-medium text-site-on-accent transition-colors hover:bg-site-accent-hover ${focusRing}`}
          >
            ۱۵ روز رایگان<span className="hidden xl:inline">&nbsp;شروع کنید</span>
          </Link>
        </div>

        <div className="flex shrink-0 items-center gap-1 lg:hidden">
          <ThemeToggle />
          <button
            type="button"
            id="nav-sheet-button"
            aria-label={sheetOpen ? 'بستن منو' : 'باز کردن منو'}
            aria-expanded={sheetOpen}
            aria-controls="nav-sheet"
            onClick={() => setSheetOpen(!sheetOpen)}
            className={`flex size-11 items-center justify-center rounded-full text-site-ink transition-colors hover:bg-site-panel ${focusRing}`}
          >
            {sheetOpen ? <X className="size-6" /> : <Menu className="size-6" />}
          </button>
        </div>
      </div>

      <div id="nav-sheet" hidden={!sheetOpen} className="fixed inset-x-0 bottom-0 top-16 flex flex-col bg-site-ground lg:hidden">
        <div className="flex-1 overflow-y-auto px-4 pb-6 pt-2">
          {navMenus.map((menu) => (
            <details key={menu.id} className="group border-b border-site-line">
              <summary className={`flex min-h-14 cursor-pointer list-none items-center justify-between rounded-xl px-2 text-base font-semibold text-site-ink ${focusRing} [&::-webkit-details-marker]:hidden`}>
                {menu.label}
                <ChevronDown aria-hidden="true" className="size-5 text-site-ink-soft transition-transform duration-300 group-open:rotate-180" />
              </summary>
              <div className="flex flex-col gap-1 pb-4">
                {visibleColumns(menu.columns).map((column, index) => (
                  <div key={column.title ?? index} className="flex flex-col gap-1">
                    {column.title ? <p className="m-0 px-3 pb-1 pt-3 text-xs font-medium text-site-ink-soft">{column.title}</p> : null}
                    {visibleLinks(column.links).map((item) => (
                      <MenuRow key={item.href} item={item} onNavigate={closeSheet} />
                    ))}
                  </div>
                ))}
              </div>
            </details>
          ))}
          <Link
            href={pricingLink.href}
            onClick={closeSheet}
            className={`flex min-h-14 items-center border-b border-site-line px-2 text-base font-semibold text-site-ink ${focusRing}`}
          >
            {pricingLink.label}
          </Link>
          <Link
            href="/login"
            onClick={closeSheet}
            className={`flex min-h-14 items-center px-2 text-base font-semibold text-site-ink ${focusRing}`}
          >
            ورود
          </Link>
        </div>

        <div className="grid grid-cols-2 gap-3 border-t border-site-line bg-site-ground p-4">
          <Link
            href="/demo"
            onClick={closeSheet}
            className={`flex min-h-12 items-center justify-center rounded-full border border-site-line px-4 text-sm font-medium text-site-ink ${focusRing}`}
          >
            درخواست دمو
          </Link>
          <Link
            href="/onboarding"
            onClick={closeSheet}
            className={`flex min-h-12 items-center justify-center rounded-full bg-site-accent px-4 text-center text-sm font-medium text-site-on-accent ${focusRing}`}
          >
            ۱۵ روز رایگان شروع کنید
          </Link>
        </div>
      </div>
    </nav>
  );
}
