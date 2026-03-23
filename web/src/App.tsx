import React, { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import AppLayout from './layouts/AppLayout'
import StandaloneLayout from './layouts/StandaloneLayout'

function DarkModeToggle() {
  const [isDark, setIsDark] = useState(false);
  
  useEffect(() => {
    if (localStorage.getItem('theme') === 'dark' || (!('theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
      document.documentElement.classList.add('dark');
      setIsDark(true);
    } else {
      document.documentElement.classList.remove('dark');
      setIsDark(false);
    }
  }, []);

  const toggleTheme = () => {
    if (isDark) {
      document.documentElement.classList.remove('dark');
      localStorage.setItem('theme', 'light');
      setIsDark(false);
    } else {
      document.documentElement.classList.add('dark');
      localStorage.setItem('theme', 'dark');
      setIsDark(true);
    }
  };

  return (
    <button 
      onClick={toggleTheme} 
      className="fixed bottom-24 md:bottom-6 right-6 p-4 rounded-full bg-primary dark:bg-[#f5f0e7] text-surface dark:text-primary shadow-2xl z-[100] transition-all hover:scale-110 active:scale-95 border border-primary/10"
      aria-label="Toggle Dark Mode"
    >
      <span className="material-symbols-outlined">{isDark ? 'light_mode' : 'dark_mode'}</span>
    </button>
  );
}

import Onboarding from './pages/Onboarding'
import Home from './pages/Home'





function App() {
  return (
    <>
      <BrowserRouter>
        <Routes>
          {/* Main App Routes (Includes Navigation) */}
          <Route element={<AppLayout />}>
            <Route path="/" element={<Home />} />
            <Route path="/vault" element={<div className="p-8 text-2xl font-headline text-center mt-20">Treasury / Vault...</div>} />
            <Route path="/notifications" element={<div className="p-8 text-2xl font-headline text-center mt-20">Oracle synthesis...</div>} />
            <Route path="/settings" element={<div className="p-8 text-2xl font-headline text-center mt-20">Control panel...</div>} />
          </Route>

          {/* Standalone Immersive Routes */}
          <Route element={<StandaloneLayout />}>
            <Route path="/onboarding" element={<Onboarding />} />
            <Route path="/checkout" element={<div className="flex items-center justify-center h-screen"><h1 className="text-4xl text-primary dark:text-white">Checkout Drawer Segment</h1></div>} />
          </Route>
        </Routes>
      </BrowserRouter>
      {/* 全局悬浮组件 */}
      <DarkModeToggle />
    </>
  )
}

export default App
