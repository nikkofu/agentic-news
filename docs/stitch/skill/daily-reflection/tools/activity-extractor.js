#!/usr/bin/env node
/**
 * Daily Reflection Tool: Activity Extractor
 * An OpenClaw AgentSkill Tool.
 * Parses the local environment (e.g., git commits, bash history, OpenClaw memories) 
 * for the current day to provide context to the Oracle (Agent).
 */
import { execSync } from 'child_process';
import fs from 'fs';
import path from 'path';

function getGitCommitsToday() {
  try {
    // Attempt to read git logs for today from the current context
    const output = execSync('git log --since="midnight" --oneline', { encoding: 'utf-8', stdio: ['pipe', 'pipe', 'ignore'] });
    return output.trim() ? output : 'No git commits today.';
  } catch (e) {
    return 'Git repository not found or no commits.';
  }
}

function getAgentMemoryToday() {
  const dateStr = new Date().toISOString().split('T')[0];
  const memoryPath = path.join(process.env.HOME || '', '.openclaw', 'workspace', 'memory', `${dateStr}.md`);
  if (fs.existsSync(memoryPath)) {
    return fs.readFileSync(memoryPath, 'utf-8');
  }
  return 'No distinct agent memory files found for today.';
}

function main() {
  const activityReport = {
    timestamp: new Date().toISOString(),
    gitActivity: getGitCommitsToday(),
    agentMemories: getAgentMemoryToday(),
    // Future expansion: hook into Safari/Chrome sqlite DB for domains visited,
    // or VS Code usage metrics via API.
  };

  // Output JSON schema directly to stdout so the LLM Agent can parse it cleanly
  console.log(JSON.stringify(activityReport, null, 2));
}

main();
