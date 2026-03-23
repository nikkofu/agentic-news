import React from 'react';
import { Link } from 'react-router-dom';

export default function Onboarding() {
  return (
    <>
      <header className="fixed top-0 left-0 w-full z-50">
          <div className="flex justify-between items-center px-8 py-6 w-full max-w-7xl mx-auto">
              <span className="text-xl font-headline italic text-primary dark:text-[#f5f0e7]">The Initiation</span>
              <div className="flex items-center gap-4">
                  <span className="text-xs font-label uppercase tracking-widest text-primary/40 dark:text-white/40">Step 01 / 03</span>
              </div>
          </div>
      </header>
      <div className="relative flex flex-col items-center justify-center min-h-screen px-6 pt-24 pb-32">
          {/* Background Atmosphere */}
          <div className="fixed inset-0 pointer-events-none overflow-hidden bg-background dark:bg-ds-dark-surface">
              <div className="absolute -top-1/4 -right-1/4 w-[800px] h-[800px] bg-primary/5 dark:bg-white/5 rounded-full blur-[120px] opacity-30"></div>
              <div className="absolute -bottom-1/4 -left-1/4 w-[600px] h-[600px] bg-secondary-container/20 dark:bg-secondary/20 rounded-full blur-[100px] opacity-20"></div>
          </div>
          <div className="w-full max-w-4xl space-y-12 relative z-10">
              {/* Step Indicators */}
              <div className="flex justify-center items-center gap-3">
                  <div className="h-1 w-12 rounded-full bg-primary dark:bg-white"></div>
                  <div className="h-1 w-12 rounded-full bg-primary/10 dark:bg-white/10"></div>
                  <div className="h-1 w-12 rounded-full bg-primary/10 dark:bg-white/10"></div>
              </div>
              {/* Page Header */}
              <div className="text-center space-y-4">
                  <h1 className="text-5xl md:text-7xl font-headline text-primary dark:text-[#f5f0e7] leading-tight">Define Your Sanctuary</h1>
                  <p className="text-lg text-primary/70 dark:text-[#d1d5db] max-w-xl mx-auto font-light leading-relaxed">
                      Select the domains of knowledge you wish to cultivate. This choice will shape the architecture of your Digital Sanctuary.
                  </p>
              </div>
              {/* Step 1: Domain Selection Bento Grid */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  {/* Card 1 */}
                  <div className="bg-white/40 dark:bg-white/5 p-8 rounded-xl cursor-pointer group hover:bg-primary/5 dark:hover:bg-white/10 transition-all duration-500 backdrop-blur-md border border-primary/5 dark:border-white/10">
                      <div className="mb-6 flex justify-between items-start">
                          <span className="material-symbols-outlined text-3xl text-primary/60 dark:text-white/60 group-hover:text-primary dark:group-hover:text-white transition-colors">psychology</span>
                          <span className="material-symbols-outlined text-primary/20 dark:text-white/20">add_circle</span>
                      </div>
                      <h3 className="text-2xl font-headline mb-2 text-primary dark:text-white">Artificial Intelligence</h3>
                      <p className="text-sm text-primary/50 dark:text-white/50 leading-relaxed font-light">Deep neural networks, ethics of automation, and future frontiers.</p>
                  </div>
                  {/* Card 2 (Active/Selected State) */}
                  <div className="bg-primary dark:bg-[#f5f0e7] p-8 rounded-xl cursor-pointer shadow-2xl transition-all duration-500 border border-primary/5 dark:border-white/10">
                      <div className="mb-6 flex justify-between items-start">
                          <span className="material-symbols-outlined text-3xl text-surface dark:text-primary">balance</span>
                          <span className="material-symbols-outlined text-surface dark:text-primary" style={{fontVariationSettings: "'FILL' 1"}}>check_circle</span>
                      </div>
                      <h3 className="text-2xl font-headline text-surface dark:text-primary mb-2">Stoic Philosophy</h3>
                      <p className="text-sm text-surface/70 dark:text-primary/70 leading-relaxed font-light">Ancient wisdom for modern resilience and clarity of mind.</p>
                  </div>
                  {/* Card 3 */}
                  <div className="bg-white/40 dark:bg-white/5 p-8 rounded-xl cursor-pointer group hover:bg-primary/5 dark:hover:bg-white/10 transition-all duration-500 backdrop-blur-md border border-primary/5 dark:border-white/10">
                      <div className="mb-6 flex justify-between items-start">
                          <span className="material-symbols-outlined text-3xl text-primary/60 dark:text-white/60 group-hover:text-primary dark:group-hover:text-white transition-colors">hub</span>
                          <span className="material-symbols-outlined text-primary/20 dark:text-white/20">add_circle</span>
                      </div>
                      <h3 className="text-2xl font-headline mb-2 text-primary dark:text-white">Systems Thinking</h3>
                      <p className="text-sm text-primary/50 dark:text-white/50 leading-relaxed font-light">Understanding the interconnected patterns of complexity.</p>
                  </div>
              </div>
          </div>
      </div>
      {/* Fixed Bottom Footer */}
      <footer className="fixed bottom-0 left-0 w-full p-8 z-50 bg-background/80 dark:bg-ds-dark-surface/80 backdrop-blur-md border-t border-primary/5 dark:border-white/5">
          <div className="max-w-7xl mx-auto flex flex-col md:flex-row justify-between items-center gap-6">
              <div className="hidden md:block">
                  <p className="text-xs font-label text-primary/30 dark:text-white/30 max-w-xs">
                      By commencing, you agree to the Digital Sanctuary protocol of mindful consumption.
                  </p>
              </div>
              <Link to="/home" className="w-full md:w-auto px-12 py-4 rounded-xl bg-primary dark:bg-[#f5f0e7] text-surface dark:text-primary font-headline text-xl shadow-2xl flex items-center justify-center gap-4 group transition-transform active:scale-[0.98] duration-200">
                  Commence
                  <span className="material-symbols-outlined group-hover:translate-x-1 transition-transform">arrow_forward</span>
              </Link>
          </div>
      </footer>
    </>
  );
}
