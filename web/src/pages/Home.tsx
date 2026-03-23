import React from 'react';
import clsx from 'clsx';
import { Link } from 'react-router-dom';

export default function Home() {
  return (
    <div className="max-w-screen-xl mx-auto px-5 w-full bg-background dark:bg-ds-dark-surface text-on-surface dark:text-ds-dark-on-surface transition-colors duration-500">
      {/* Hero Section: Today's Top Insight */}
      <section className="mb-12">
          <div className="flex flex-col gap-6">
              <div>
                  <span className="font-body text-[10px] uppercase tracking-[0.3em] text-secondary dark:text-[#b4c8e4] font-extrabold mb-3 block">Focus of the Day</span>
                  <h2 className="text-4xl md:text-6xl font-headline italic text-primary dark:text-[#f5f0e7] leading-[1.15] mb-4 font-medium">The Architecture of Silence</h2>
                  <p className="font-body text-base leading-relaxed text-on-surface-variant dark:text-white/70 mb-6 max-w-xl">
                      Explore how physical environments dictate cognitive depth. A journey through historical libraries and the modern need for "The Scholarly Atrium."
                  </p>
                  <button className="w-full md:w-fit px-8 py-4 rounded-xl bg-primary dark:bg-[#f5f0e7] text-on-primary dark:text-primary font-body font-bold flex items-center justify-center gap-3 active:scale-[0.98] transition-all shadow-lg shadow-primary/20 hover:bg-primary/90 dark:hover:bg-white">
                      Begin Deep Dive
                      <span className="material-symbols-outlined text-lg">arrow_forward</span>
                  </button>
              </div>
              
              <div className="relative group">
                  <div className="aspect-[16/10] md:aspect-[21/9] rounded-2xl overflow-hidden bg-surface-container-low dark:bg-white/5 shadow-sm border border-outline-variant/20 dark:border-white/10">
                      <img alt="Insight Visual" className="w-full h-full object-cover opacity-95 group-hover:scale-105 transition-transform duration-1000" src="https://lh3.googleusercontent.com/aida-public/AB6AXuCG28RxyymAd90zNWDNDwfAHIQCRBxFJR3q34nenXUCx49AgWyif4sb14c4d8SnwlA4QEy2m-SRVVDeFOaC6iEtroqQaf_erFDLQeV89au27_sJhtozp1BzJDiHPXj6YwHGnn-9C63Cr00NKBkY8OJo9O_HPRetQBRc_fX9EDleKI9esJOe_V5XZBxVX8EUF1Pzn8qCRoFbyd4bt3hhaBfnbidf6l_99udpUOvhjUeLDB1LgR4riZ6tFrHZrXZ1W6jL03DNzNBJEQ4" />
                      <div className="absolute inset-0 bg-gradient-to-t from-black/40 via-transparent to-transparent"></div>
                  </div>
                  {/* Quote Overlay */}
                  <div className="absolute -bottom-4 left-4 right-4 p-5 bg-white/90 dark:bg-[#030b12]/90 backdrop-blur-md rounded-xl border border-white/20 dark:border-white/10 shadow-xl hidden md:block max-w-xs transition-transform group-hover:-translate-y-2">
                      <span className="material-symbols-outlined text-primary dark:text-[#f5f0e7] mb-2 text-xl">auto_stories</span>
                      <p className="font-headline italic text-primary-container dark:text-[#f5f0e7] text-md font-semibold">"Silence is the sleep that nourishes wisdom."</p>
                      <p className="font-body text-[10px] text-secondary dark:text-white/60 mt-3 uppercase tracking-widest font-extrabold">Francis Bacon</p>
                  </div>
              </div>
          </div>
      </section>

      {/* Butler's Suggestion: Native Widget Style */}
      <section className="mb-12">
          <div className="bg-surface-container-low dark:bg-white/5 p-7 rounded-2xl border border-outline-variant/30 dark:border-white/10 shadow-sm hover:shadow-md transition-shadow">
              <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
                  <div className="flex items-center gap-4">
                      <div className="w-12 h-12 rounded-xl bg-primary/5 dark:bg-[#f5f0e7]/10 flex items-center justify-center text-primary dark:text-[#f5f0e7]">
                          <span className="material-symbols-outlined text-2xl">psychology</span>
                      </div>
                      <div>
                          <h3 className="font-headline text-lg font-bold text-primary dark:text-[#f5f0e7]">Butler's Suggestion</h3>
                          <p className="font-body text-xs text-on-surface-variant dark:text-white/60">Endurance is at <span className="text-primary dark:text-white font-bold">82%</span></p>
                      </div>
                  </div>
                  <div className="text-left sm:text-right w-full sm:w-auto flex justify-between sm:block border-t border-outline-variant/20 dark:border-white/10 pt-4 sm:pt-0">
                      <p className="text-[10px] font-bold text-on-surface-variant dark:text-white/60 uppercase tracking-widest">Progress</p>
                      <p className="text-sm font-headline font-bold text-primary dark:text-[#f5f0e7]">45 / 60 <span className="text-[10px] font-body text-secondary dark:text-white/40">min</span></p>
                  </div>
              </div>
              <div className="space-y-6">
                  <div className="h-2 w-full bg-surface-container-highest dark:bg-black/40 rounded-full overflow-hidden shadow-inner flex">
                      <div className="h-full bg-primary dark:bg-white/80 w-[75%] rounded-full relative">
                          <div className="absolute inset-0 bg-white/20 w-full animate-pulse"></div>
                      </div>
                  </div>
                  <button className="w-full bg-primary dark:bg-[#f5f0e7] text-on-primary dark:text-primary py-3.5 rounded-xl font-body font-bold text-sm hover:bg-primary-container dark:hover:bg-white transition-colors shadow-sm">
                      Optimize My Plan
                  </button>
              </div>
          </div>
      </section>

      {/* Curated For You Feed */}
      <section>
          <div className="flex justify-between items-end mb-6">
              <h2 className="text-2xl font-headline font-bold text-primary dark:text-[#f5f0e7]">Curated for You</h2>
              <a className="font-body text-[10px] font-extrabold text-secondary dark:text-white/60 uppercase tracking-widest hover:text-primary dark:hover:text-[#f5f0e7] transition-colors border-b border-secondary/20 hover:border-primary/40 pb-0.5" href="#">See All</a>
          </div>
          
          <div className="flex flex-col gap-6">
              {/* Article Card 1 */}
              <div className="group bg-surface-container-lowest dark:bg-[#030b12] rounded-2xl overflow-hidden border border-surface-container-highest dark:border-white/10 shadow-sm hover:shadow-md transition-all cursor-pointer">
                  <div className="relative aspect-[16/9] sm:aspect-[2/1] md:aspect-[21/9] overflow-hidden">
                      <img alt="AI Ethics" className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-1000" src="https://lh3.googleusercontent.com/aida-public/AB6AXuDk5vjpwpqhjFXeblsqiBexrGUpkmyGdsWksJJ6OT-1vNNclRRbT9ZKOnCTCnHM3elmy4eFCK8nnq3PSsJuHu_ykV_F2LmsqK5qRh2wLZQGwTvbStReFa0I1dMfjLWRJXqW-OtrDQs7W7dhgAElSyeFda034boysNm7q3WD229Uv7sXyzF7kZ2fI-Chr6lKcOpl7WUyNsz0u_8YWPCAbtL-wx0tJ5M11LbQ7q9qtkX5StpD2QdDvzCY1MKmzdP1gFljTgkGbLoIf4c" />
                      <div className="absolute top-4 left-4">
                          <span className="px-3 py-1 rounded-lg bg-black/40 backdrop-blur-md text-white text-[9px] font-bold uppercase tracking-widest shadow-sm">Ethics &amp; Frontier</span>
                      </div>
                      <button className="absolute top-4 right-4 w-10 h-10 rounded-full bg-white/10 backdrop-blur-md text-white border border-white/20 flex items-center justify-center hover:bg-white/30 transition-colors shadow-sm active:scale-95">
                          <span className="material-symbols-outlined text-xl">bookmark</span>
                      </button>
                  </div>
                  <div className="p-6 md:p-8">
                      <h3 className="text-xl md:text-2xl font-headline font-bold text-primary dark:text-[#f5f0e7] leading-tight mb-3 group-hover:text-secondary dark:group-hover:text-white transition-colors">The Moral Ghost in the Machine</h3>
                      <p className="font-body text-sm text-on-surface-variant dark:text-white/60 leading-relaxed line-clamp-2 md:line-clamp-3 mb-6">Analyzing the alignment problem in generative models through the lens of Kantian ethics, exploring whether artificial agents can truly possess moral agency or if they are simply sophisticated mirrors of human values.</p>
                      
                      <div className="flex items-center justify-between border-t border-surface-container-high dark:border-white/10 pt-4">
                          <div className="flex items-center gap-4 text-on-surface-variant/80 dark:text-white/50">
                              <div className="flex items-center gap-1.5">
                                  <span className="material-symbols-outlined text-lg">schedule</span>
                                  <span className="text-xs font-bold font-body">12 min</span>
                              </div>
                              <div className="flex items-center gap-1.5">
                                  <span className="material-symbols-outlined text-lg text-secondary dark:text-white/80">local_library</span>
                                  <span className="text-xs font-bold font-body text-secondary dark:text-white/80">Deep Focus</span>
                              </div>
                          </div>
                          <span className="material-symbols-outlined text-primary dark:text-[#f5f0e7] group-hover:translate-x-1 transition-transform">arrow_forward</span>
                      </div>
                  </div>
              </div>

              {/* Horizontal Brief Cards Scroll */}
              <div className="mt-2 mb-2">
                  <div className="flex justify-between items-center mb-4">
                      <span className="font-body text-[10px] uppercase tracking-widest font-bold text-outline dark:text-white/40">Quick Reads</span>
                  </div>
                  <div className="flex overflow-x-auto gap-4 custom-scrollbar pb-4 -mx-5 px-5 md:mx-0 md:px-0">
                      {/* Brief Card 1 */}
                      <div className="min-w-[85%] sm:min-w-[320px] bg-secondary-container/50 dark:bg-white/5 hover:bg-secondary-container dark:hover:bg-white/10 rounded-2xl p-6 border border-secondary/10 dark:border-white/5 flex flex-col justify-between cursor-pointer transition-colors group shadow-sm">
                          <div>
                              <div className="flex justify-between items-start mb-4">
                                  <div className="w-10 h-10 rounded-full bg-white/50 dark:bg-black/30 flex items-center justify-center shadow-sm">
                                      <span className="material-symbols-outlined text-xl text-on-secondary-container dark:text-white/70">trending_up</span>
                                  </div>
                                  <span className="text-[9px] font-extrabold uppercase tracking-widest text-on-secondary-container/60 dark:text-white/50 bg-white/30 dark:bg-black/30 px-2 py-1 rounded-md">Macro-Trends</span>
                              </div>
                              <h3 className="text-lg font-headline font-bold text-primary dark:text-[#f5f0e7] mb-2 group-hover:text-on-secondary-container transition-colors">Resilience: Friend-Shoring</h3>
                              <p className="font-body text-[13px] text-on-surface-variant dark:text-white/60 leading-relaxed">How global supply shifts impact local sustainability efforts and reshape economic alliances.</p>
                          </div>
                          <div className="mt-6 flex items-center justify-between pt-4 border-t border-secondary/20 dark:border-white/10">
                              <span className="text-[11px] font-bold text-on-secondary-container dark:text-[#f5f0e7] flex items-center gap-1 group-hover:gap-2 transition-all">
                                  Read Brief <span className="material-symbols-outlined text-[14px]">east</span>
                              </span>
                              <span className="text-[10px] font-bold text-secondary/60 dark:text-white/40 bg-white/40 dark:bg-black/40 px-2 py-0.5 rounded">4 min</span>
                          </div>
                      </div>
                      
                      {/* Brief Card 2 */}
                      <div className="min-w-[85%] sm:min-w-[320px] bg-surface-container dark:bg-white/5 hover:bg-surface-container-high dark:hover:bg-white/10 rounded-2xl p-6 border border-outline-variant/10 dark:border-white/5 flex flex-col justify-between cursor-pointer transition-colors group shadow-sm">
                          <div>
                              <div className="flex justify-between items-start mb-4">
                                  <div className="w-10 h-10 rounded-full bg-white dark:bg-black/30 flex items-center justify-center shadow-sm">
                                      <span className="material-symbols-outlined text-xl text-primary dark:text-[#f5f0e7]">history_edu</span>
                                  </div>
                                  <span className="text-[9px] font-extrabold uppercase tracking-widest text-secondary dark:text-[#f5f0e7] bg-white dark:bg-black/50 px-2 py-1 rounded-md shadow-sm border border-outline-variant/10 dark:border-white/10">Modern Classics</span>
                              </div>
                              <h3 className="text-lg font-headline font-bold text-primary dark:text-[#f5f0e7] mb-2">Digital Stoicism</h3>
                              <p className="font-body text-[13px] text-on-surface-variant dark:text-white/60 leading-relaxed">Applying principles from Marcus Aurelius to navigate the modern economy of attention.</p>
                          </div>
                          <div className="mt-6 flex items-center justify-between pt-4 border-t border-outline-variant/20 dark:border-white/10">
                              <span className="text-[11px] font-bold text-primary dark:text-[#f5f0e7] flex items-center gap-1 group-hover:gap-2 transition-all">
                                  Read Brief <span className="material-symbols-outlined text-[14px]">east</span>
                              </span>
                              <span className="text-[10px] font-bold text-outline dark:text-white/40 bg-white dark:bg-black/30 px-2 py-0.5 rounded shadow-sm">6 min</span>
                          </div>
                      </div>
                  </div>
              </div>

              {/* Premium Nudge */}
              <div className="mt-4 bg-primary dark:bg-[#030b12] text-on-primary dark:text-[#f5f0e7] rounded-3xl p-8 md:p-10 relative overflow-hidden shadow-xl shadow-primary/20 dark:border dark:border-white/10 scholarly-gradient">
                  <div className="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-8">
                      <div className="max-w-md">
                          <div className="flex items-center gap-2 mb-4">
                              <span className="material-symbols-outlined text-xl text-tertiary-fixed dark:text-white">auto_awesome</span>
                              <span className="text-[10px] font-extrabold uppercase tracking-[0.2em] text-on-primary-container dark:text-white/80 bg-black/20 dark:bg-white/10 px-3 py-1 rounded-full backdrop-blur-sm border border-white/10">Library Expansion</span>
                          </div>
                          <h3 className="text-2xl md:text-3xl font-headline font-bold leading-tight mb-3">Unlock the Sanctuary</h3>
                          <p className="text-sm text-on-primary-container dark:text-white/60 leading-relaxed opacity-90 font-light">Gain exclusive access to high-fidelity audio archives, historical manuscripts, and unlimited focus sessions.</p>
                      </div>
                      <div className="flex flex-col sm:flex-row items-center gap-6 shrink-0 z-20">
                          <button className="w-full sm:w-auto bg-surface dark:bg-[#f5f0e7] text-primary dark:text-primary px-8 py-3.5 rounded-xl font-bold text-sm shadow-lg hover:bg-surface-container-high transition-colors active:scale-95 text-center">
                              Upgrade Now
                          </button>
                      </div>
                  </div>
                  {/* Subtle background decoration */}
                  <div className="absolute -right-10 -bottom-10 w-64 h-64 bg-white/5 rounded-full blur-3xl pointer-events-none"></div>
                  <div className="absolute -top-20 -left-10 w-40 h-40 bg-secondary/10 dark:bg-white/5 rounded-full blur-2xl pointer-events-none"></div>
                  {/* Large icon in background */}
                  <span className="material-symbols-outlined text-white/5 text-[120px] absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none hidden md:block" style={{ fontVariationSettings: "'wght' 200" }}>workspace_premium</span>
              </div>
          </div>
      </section>
    </div>
  );
}
