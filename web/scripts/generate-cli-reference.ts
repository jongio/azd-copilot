/**
 * CLI Reference Generator for azd-copilot
 * 
 * Generates reference pages from the CLI help output at build time.
 * Creates:
 * - /reference/cli/index.astro (overview)
 * - /reference/cli/[command].astro (individual command pages)
 */

import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';
import { execSync } from 'child_process';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const OUTPUT_DIR = path.resolve(__dirname, '../src/pages/reference/cli');

interface CommandInfo {
  name: string;
  description: string;
  usage: string;
  flags: Flag[];
}

interface Flag {
  flag: string;
  short: string;
  description: string;
}

/**
 * Parse the help output from a command
 */
function parseHelpOutput(helpText: string): { description: string; flags: Flag[] } {
  const lines = helpText.split('\n');
  let description = '';
  const flags: Flag[] = [];
  
  // First non-empty line is usually the description
  for (const line of lines) {
    if (line.trim() && !line.startsWith('Usage:') && !line.startsWith('Available Commands:') && !line.startsWith('Flags:')) {
      description = line.trim();
      break;
    }
  }
  
  // Parse flags section
  let inFlags = false;
  for (const line of lines) {
    if (line.startsWith('Flags:')) {
      inFlags = true;
      continue;
    }
    if (line.startsWith('Use "') || (inFlags && line.trim() === '')) {
      inFlags = false;
      continue;
    }
    if (inFlags && line.trim()) {
      // Parse flag line like: "  -h, --help   help for version"
      const match = line.match(/^\s+(-\w)?,?\s*(--[\w-]+)\s+(.+)$/);
      if (match) {
        flags.push({
          short: match[1] || '',
          flag: match[2],
          description: match[3].trim()
        });
      } else {
        // Try matching just long flag
        const longMatch = line.match(/^\s+(--[\w-]+)\s+(.+)$/);
        if (longMatch) {
          flags.push({
            short: '',
            flag: longMatch[1],
            description: longMatch[2].trim()
          });
        }
      }
    }
  }
  
  return { description, flags };
}

/**
 * Get commands by running the CLI
 */
function discoverCommands(): CommandInfo[] {
  const commands: CommandInfo[] = [];
  
  try {
    // Get main help
    const mainHelp = execSync('go run ./src/cmd/copilot --help 2>&1', { 
      encoding: 'utf-8',
      cwd: path.resolve(__dirname, '../..')
    });
    
    // Find available commands
    const commandMatch = mainHelp.match(/Available Commands:\n([\s\S]*?)(?=\n\nFlags:|$)/);
    if (commandMatch) {
      const commandLines = commandMatch[1].split('\n').filter(l => l.trim());
      
      for (const line of commandLines) {
        const match = line.match(/^\s+(\w+)\s+(.+)$/);
        if (match && !['completion', 'help'].includes(match[1])) {
          const cmdName = match[1];
          const cmdDesc = match[2].trim();
          
          // Get detailed help for this command
          try {
            const cmdHelp = execSync(`go run ./src/cmd/copilot ${cmdName} --help 2>&1`, {
              encoding: 'utf-8',
              cwd: path.resolve(__dirname, '../..')
            });
            
            const parsed = parseHelpOutput(cmdHelp);
            commands.push({
              name: cmdName,
              description: cmdDesc,
              usage: `azd copilot ${cmdName} [flags]`,
              flags: parsed.flags
            });
          } catch {
            commands.push({
              name: cmdName,
              description: cmdDesc,
              usage: `azd copilot ${cmdName} [flags]`,
              flags: []
            });
          }
        }
      }
    }
  } catch (err) {
    console.warn('⚠️  Could not run CLI to discover commands. Skipping generation to preserve existing pages.');
    return commands;
  }
  
  return commands;
}

function generateFlagsTable(command: CommandInfo): string {
  if (command.flags.length === 0) return '';
  
  const rows = command.flags.map(f => 
    `<tr class="border-t border-[var(--color-border)]">
      <td class="py-3 px-4"><code>${f.flag}</code></td>
      <td class="py-3 px-4">${f.short ? `<code>${f.short}</code>` : '-'}</td>
      <td class="py-3 px-4">${f.description}</td>
    </tr>`
  ).join('\n');
  
  return `
<h2>Flags</h2>
<div class="overflow-x-auto my-4">
  <table class="min-w-full text-sm">
    <thead>
      <tr class="bg-[var(--color-muted)]">
        <th class="text-left py-3 px-4 font-semibold">Flag</th>
        <th class="text-left py-3 px-4 font-semibold">Short</th>
        <th class="text-left py-3 px-4 font-semibold">Description</th>
      </tr>
    </thead>
    <tbody>
      ${rows}
    </tbody>
  </table>
</div>`;
}

function generateCommandPage(command: CommandInfo): string {
  return `---
import Layout from '../../../components/Layout.astro';
---

<Layout title="${command.name} - CLI Reference">
  <div class="content">
    <nav class="breadcrumb">
      <a href="/azd-copilot/">Home</a> /
      <a href="/azd-copilot/reference/cli/">CLI Reference</a> /
      <span>${command.name}</span>
    </nav>

    <h1>azd copilot ${command.name}</h1>
    <p class="description">${command.description}</p>

    <h2>Usage</h2>
    <pre><code>${command.usage}</code></pre>

    ${generateFlagsTable(command)}

    <div class="back-link">
      <a href="/azd-copilot/reference/cli/">← Back to CLI Reference</a>
    </div>
  </div>
</Layout>

<style>
  .content {
    max-width: 48rem;
    margin: 0 auto;
    padding: 3rem 1.5rem;
  }

  .breadcrumb {
    font-size: 0.875rem;
    color: var(--color-muted-foreground);
    margin-bottom: 2rem;
  }

  .breadcrumb a {
    color: var(--color-muted-foreground);
  }

  .breadcrumb a:hover {
    color: var(--color-primary);
  }

  h1 {
    margin-bottom: 1rem;
  }

  .description {
    font-size: 1.125rem;
    color: var(--color-muted-foreground);
    margin-bottom: 2rem;
  }

  h2 {
    font-size: 1.5rem;
    margin-top: 2rem;
    margin-bottom: 1rem;
  }

  .back-link {
    margin-top: 3rem;
    padding-top: 2rem;
    border-top: 1px solid var(--color-border);
  }
</style>
`;
}

function generateIndexPage(commands: CommandInfo[]): string {
  const commandCards = commands.map(cmd => `
    <a href="/azd-copilot/reference/cli/${cmd.name}/" class="command-card">
      <code class="command-name">azd copilot ${cmd.name}</code>
      <p class="command-desc">${cmd.description}</p>
      <span class="command-meta">${cmd.flags.length} flags</span>
    </a>
  `).join('\n');

  return `---
import Layout from '../../../components/Layout.astro';
---

<Layout title="CLI Reference">
  <div class="content">
    <h1>CLI Reference</h1>
    <p class="intro">
      Complete reference for all <code>azd copilot</code> commands and flags.
    </p>

    <section>
      <h2>Global Flags</h2>
      <p>These flags are available for all commands:</p>
      <div class="overflow-x-auto">
        <table class="flags-table">
          <thead>
            <tr>
              <th>Flag</th>
              <th>Short</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr><td><code>--output</code></td><td><code>-o</code></td><td>Output format (default, json)</td></tr>
            <tr><td><code>--debug</code></td><td>-</td><td>Enable debug logging</td></tr>
            <tr><td><code>--structured-logs</code></td><td>-</td><td>Enable structured JSON logging to stderr</td></tr>
            <tr><td><code>--cwd</code></td><td><code>-C</code></td><td>Sets the current working directory</td></tr>
          </tbody>
        </table>
      </div>
    </section>

    <section>
      <h2>Commands</h2>
      <div class="commands-grid">
        ${commandCards}
      </div>
    </section>
  </div>
</Layout>

<style>
  .content {
    max-width: 64rem;
    margin: 0 auto;
    padding: 3rem 1.5rem;
  }

  h1 {
    margin-bottom: 1rem;
  }

  .intro {
    font-size: 1.25rem;
    color: var(--color-muted-foreground);
    margin-bottom: 3rem;
  }

  section {
    margin-bottom: 3rem;
  }

  h2 {
    font-size: 1.5rem;
    margin-bottom: 1rem;
  }

  .flags-table {
    width: 100%;
    font-size: 0.875rem;
    border: 1px solid var(--color-border);
    border-radius: 0.5rem;
    overflow: hidden;
  }

  .flags-table th {
    text-align: left;
    padding: 0.75rem 1rem;
    background: var(--color-muted);
    font-weight: 600;
  }

  .flags-table td {
    padding: 0.75rem 1rem;
    border-top: 1px solid var(--color-border);
  }

  .commands-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1rem;
  }

  .command-card {
    display: block;
    padding: 1.5rem;
    background: var(--color-secondary);
    border: 1px solid var(--color-border);
    border-radius: 0.5rem;
    text-decoration: none;
    transition: border-color 0.15s ease;
  }

  .command-card:hover {
    border-color: var(--color-primary);
    text-decoration: none;
  }

  .command-name {
    display: block;
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-primary);
    margin-bottom: 0.5rem;
  }

  .command-desc {
    color: var(--color-muted-foreground);
    margin: 0 0 0.75rem 0;
    font-size: 0.875rem;
  }

  .command-meta {
    font-size: 0.75rem;
    color: var(--color-muted-foreground);
  }
</style>
`;
}

async function main() {
  console.log('🔧 Generating CLI reference pages...\n');
  
  // Discover commands
  const commands = discoverCommands();
  
  if (commands.length === 0) {
    console.log('  ⏭️  No commands discovered. Preserving existing pages.\n');
    return;
  }
  
  console.log(`  📋 Discovered ${commands.length} commands: ${commands.map(c => c.name).join(', ')}\n`);
  
  // Ensure output directory exists
  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  }
  
  // Generate index page
  const indexPage = generateIndexPage(commands);
  fs.writeFileSync(path.join(OUTPUT_DIR, 'index.astro'), indexPage);
  console.log(`  ✓ Generated: reference/cli/index.astro`);
  
  // Generate individual command pages
  for (const cmd of commands) {
    const page = generateCommandPage(cmd);
    fs.writeFileSync(path.join(OUTPUT_DIR, `${cmd.name}.astro`), page);
    console.log(`  ✓ Generated: reference/cli/${cmd.name}.astro`);
  }
  
  console.log(`\n✅ Generated ${commands.length + 1} CLI reference pages`);
}

main().catch(err => {
  console.error('Error generating CLI reference:', err);
  process.exit(1);
});
