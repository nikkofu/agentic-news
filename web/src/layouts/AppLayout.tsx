import React from 'react';
import { Outlet, Link, useLocation } from 'react-router-dom';
import clsx from 'clsx';

export default function AppLayout() {
  const location = useLocation();

  const navItems = [
    { name: 'Sanctuary', path: '/', icon: 'auto_stories' },
    { name: 'Treasury', path: '/vault', icon: 'payments' },
    { name: 'Oracle', path: '/notifications', icon: 'auto_awesome' },
    { name: 'Control', path: '/settings', icon: 'settings_input_component' }
  ];

  return (
    <div className="bg-surface dark:bg-ds-dark-surface min-h-screen text-on-surface dark:text-ds-dark-on-surface transition-colors duration-500 font-body">
      {/* Top App Bar (Desktop) */}
      <header className="hidden md:flex fixed top-0 left-0 w-full bg-primary/95 dark:bg-[#030b12]/95 backdrop-blur-sm text-surface z-50 items-center px-8 py-4 border-b border-primary-container/20">
        <div className="max-w-7xl mx-auto w-full flex justify-between items-center">
          <div className="flex items-center gap-4">
            <span className="material-symbols-outlined text-surface">menu</span>
            <span className="text-xl font-headline italic tracking-wide">The Digital Sanctuary</span>
          </div>
          <div className="flex items-center gap-12">
            <nav className="flex items-center gap-8 text-sm font-semibold tracking-wide">
              {navItems.map(item => (
                <Link
                  key={item.path}
                  to={item.path}
                  className={clsx(
                    "transition-colors",
                    location.pathname === item.path ? "text-surface border-b-2 border-[#b08d1a] pb-1" : "text-surface/60 hover:text-surface"
                  )}
                >
                  {item.name}
                </Link>
              ))}
            </nav>
            <div className="h-10 w-10 bg-surface/10 rounded-full border border-surface/20 overflow-hidden shadow-inner flex items-center justify-center">
               <span className="material-symbols-outlined text-surface/50">person</span>
            </div>
          </div>
        </div>
      </header>
      
      {/* Spacer for Top Header on Desktop */}
      <div className="hidden md:block h-20 w-full"></div>

      {/* Main Content Area */}
      <main className="max-w-7xl mx-auto w-full pb-28 md:pb-12 pt-8">
        <Outlet />
      </main>

      {/* Bottom Navigation (Mobile) */}
      <nav className="md:hidden fixed bottom-0 left-0 w-full z-50 bg-primary dark:bg-[#030b12] flex justify-around items-center px-6 pb-6 pt-3 shadow-[0_-4px_24px_rgba(0,0,0,0.2)]">
        {navItems.map(item => {
          const isActive = location.pathname === item.path;
          return (
            <Link
              key={item.path}
              to={item.path}
              className={clsx(
                "flex flex-col items-center justify-center px-4 py-2 transition-all",
                isActive ? "bg-primary-container dark:bg-white/10 text-surface rounded-xl px-5 py-2.5 -translate-y-2 shadow-lg" : "text-surface opacity-60 hover:opacity-100"
              )}
            >
              <span className="material-symbols-outlined mb-1 text-2xl" style={{ fontVariationSettings: isActive ? "'FILL' 1" : "'FILL' 0" }}>
                {item.icon}
              </span>
              <span className="text-[10px] font-bold tracking-widest uppercase mt-1">{item.name}</span>
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
