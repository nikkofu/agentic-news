#!/usr/bin/env node
/**
 * Daily Reflection Tool: Attention & Entropy Analyzer
 * Computes "deep work" vs "context switching entropy" by analyzing 
 * timestamp gaps and commit variance across the active workspace.
 */
import { execSync } from 'child_process';

function analyzeGitRhythm() {
  try {
    // 提取今日 6am 以后的所有 commit UNIX 时间戳
    const log = execSync("git log --since='6am' --format='%cd' --date=unix", { encoding: 'utf-8', stdio: ['pipe', 'pipe', 'ignore'] }).trim();
    if (!log) return { focusScore: 0, entropy: 'Low', status: 'Dormant (Observer)' };
    
    // 自研简单启发式算法：若 commit 间隔太近，视为高频摩擦/试错；
    // 若呈现 45-120 分钟的间断性规整打包，则判定为处于深度专注心流状态 (Deep Work Block)。
    const timestamps = log.split('\n').map(Number).sort((a,b) => a - b);
    let deepWorkBlocks = 0;
    
    for (let i = 1; i < timestamps.length; i++) {
        const gap = (timestamps[i] - timestamps[i-1]) / 60; // 转为 minutes
        if (gap > 45 && gap < 120) deepWorkBlocks++;
    }
    
    return {
      focusScore: Math.min(100, (deepWorkBlocks + 1) * 20 + timestamps.length * 2),
      entropy: timestamps.length > 15 ? 'High (Scattered Tasking)' : (timestamps.length <= 4 ? 'Low (Laser Focused)' : 'Medium'),
      status: deepWorkBlocks >= 2 ? 'Flow State Achieved' : 'Shallow Operating Mode'
    };
  } catch (e) {
    return { focusScore: null, error: 'No Git contextual telemetry found' };
  }
}

console.log(JSON.stringify(analyzeGitRhythm(), null, 2));
