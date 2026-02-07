// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package scenario

import (
	"fmt"
	"os"
)

// GenerateDashboard creates a self-contained HTML dashboard that loads
// results.db via sql.js (WASM SQLite) directly in the browser.
// The dashboard reads the DB live — no regeneration needed after new runs.
func GenerateDashboard(db *DB, outPath string) error {
	_ = db // DB is not read here — the HTML loads it client-side
	return os.WriteFile(outPath, []byte(dashboardHTML), 0644)
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>azd-copilot Scenario Dashboard</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/sql.js/1.10.3/sql-wasm.js"></script>
<style>
  :root {
    --bg: #0d1117; --surface: #161b22; --border: #30363d;
    --text: #e6edf3; --text-dim: #8b949e; --accent: #58a6ff;
    --green: #3fb950; --red: #f85149; --yellow: #d29922; --purple: #bc8cff;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif;
    background: var(--bg); color: var(--text); line-height: 1.5; padding: 24px; }
  h1 { font-size: 28px; margin-bottom: 4px; }
  .subtitle { color: var(--text-dim); font-size: 14px; margin-bottom: 24px; }
  .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 16px; margin-bottom: 32px; }
  .card { background: var(--surface); border: 1px solid var(--border); border-radius: 8px; padding: 20px; }
  .card-value { font-size: 36px; font-weight: 700; }
  .card-label { color: var(--text-dim); font-size: 13px; text-transform: uppercase; letter-spacing: 0.5px; }
  .card-value.pass { color: var(--green); }
  .card-value.fail { color: var(--red); }
  .tabs { display: flex; gap: 0; border-bottom: 1px solid var(--border); margin-bottom: 24px; }
  .tab { padding: 8px 20px; cursor: pointer; color: var(--text-dim); border-bottom: 2px solid transparent;
    font-size: 14px; transition: all 0.2s; }
  .tab:hover { color: var(--text); }
  .tab.active { color: var(--accent); border-bottom-color: var(--accent); }
  .tab-content { display: none; }
  .tab-content.active { display: block; }
  .chart-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 32px; }
  .chart-box { background: var(--surface); border: 1px solid var(--border); border-radius: 8px; padding: 16px; }
  .chart-box h3 { font-size: 14px; color: var(--text-dim); margin-bottom: 12px; }
  @media (max-width: 900px) { .chart-grid { grid-template-columns: 1fr; } }
  table { width: 100%; border-collapse: collapse; font-size: 14px; }
  th { text-align: left; padding: 10px 12px; border-bottom: 2px solid var(--border);
    color: var(--text-dim); font-weight: 600; font-size: 12px; text-transform: uppercase; }
  td { padding: 10px 12px; border-bottom: 1px solid var(--border); }
  tr:hover td { background: rgba(88, 166, 255, 0.04); }
  .mono { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 12px; }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 12px; font-weight: 600; }
  .badge.pass { background: rgba(63, 185, 80, 0.15); color: var(--green); }
  .badge.fail { background: rgba(248, 81, 73, 0.15); color: var(--red); }
  .score-bar { display: inline-block; height: 8px; border-radius: 4px; background: var(--border); width: 60px; position: relative; vertical-align: middle; }
  .score-fill { height: 100%; border-radius: 4px; position: absolute; left: 0; top: 0; }
  .score-fill.high { background: var(--green); }
  .score-fill.mid { background: var(--yellow); }
  .score-fill.low { background: var(--red); }
  .compare-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 32px; }
  .compare-card { background: var(--surface); border: 1px solid var(--border); border-radius: 8px; padding: 20px; }
  .compare-card h3 { font-size: 16px; margin-bottom: 16px; }
  .metric-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid var(--border); }
  .metric-row:last-child { border-bottom: none; }
  .metric-label { color: var(--text-dim); }
  .metric-value { font-weight: 600; font-family: 'SF Mono', monospace; }
  .delta { font-size: 12px; margin-left: 4px; }
  .delta.better { color: var(--green); }
  .delta.worse { color: var(--red); }
  @media (max-width: 900px) { .compare-grid { grid-template-columns: 1fr; } }
  .detail-section { background: var(--surface); border: 1px solid var(--border); border-radius: 8px; padding: 16px; margin-top: 8px; }
  .detail-section h4 { font-size: 13px; color: var(--text-dim); margin-bottom: 8px; text-transform: uppercase; }
  .empty-state { text-align: center; padding: 60px 20px; color: var(--text-dim); }
  .empty-state h2 { font-size: 20px; margin-bottom: 8px; color: var(--text); }
  #loading { text-align: center; padding: 60px; color: var(--text-dim); font-size: 18px; }
  .refresh-btn { background: var(--surface); border: 1px solid var(--border); color: var(--accent);
    padding: 6px 16px; border-radius: 6px; cursor: pointer; font-size: 13px; margin-left: 12px; }
  .refresh-btn:hover { background: var(--border); }
</style>
</head>
<body>
<h1>📊 azd-copilot Scenario Dashboard <button class="refresh-btn" onclick="loadDB()">🔄 Refresh</button></h1>
<p class="subtitle">Live from results.db</p>
<div id="loading">Loading results.db...</div>
<div id="app" style="display:none"></div>

<script>
let SQL;
let db;

async function loadDB() {
  if (!SQL) {
    SQL = await initSqlJs({ locateFile: f => 'https://cdnjs.cloudflare.com/ajax/libs/sql.js/1.10.3/' + f });
  }

  // Load DB file relative to this HTML file
  try {
    const resp = await fetch('results.db');
    if (!resp.ok) throw new Error('Failed to load results.db: ' + resp.status);
    const buf = await resp.arrayBuffer();
    db = new SQL.Database(new Uint8Array(buf));
    render();
  } catch (e) {
    document.getElementById('loading').innerHTML =
      '<div class="empty-state"><h2>Cannot load results.db</h2><p>' + e.message +
      '</p><p style="margin-top:12px">Make sure results.db is in the same directory as this HTML file.<br>' +
      'If opening via file://, you may need a local server: <code>python -m http.server 8080</code></p></div>';
  }
}

function query(sql) {
  const result = db.exec(sql);
  if (!result.length) return [];
  const cols = result[0].columns;
  return result[0].values.map(row => {
    const obj = {};
    cols.forEach((c, i) => obj[c] = row[i]);
    return obj;
  });
}

function pct(v) { return Math.round(v * 100); }
function shortID(s) { return s ? s.slice(0, 8) : '?'; }
function shortCommit(s) { return s ? s.slice(0, 7) : '?'; }
function passFail(b) { return b ? '✅' : '❌'; }
function passClass(b) { return b ? 'pass' : 'fail'; }

function formatDur(sec) {
  if (sec < 60) return sec + 's';
  const m = Math.floor(sec / 60), s = sec % 60;
  if (m < 60) return m + 'm ' + s + 's';
  return Math.floor(m / 60) + 'h ' + (m % 60) + 'm';
}

function dateShort(s) {
  if (!s) return '?';
  const d = new Date(s);
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) + ' ' +
         d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false });
}

function scoreClass(v) { return v >= 0.8 ? 'high' : v >= 0.5 ? 'mid' : 'low'; }

function deltaHTML(newVal, oldVal, lowerIsBetter) {
  if (newVal === oldVal) return '';
  const better = lowerIsBetter ? newVal < oldVal : newVal > oldVal;
  return ' <span class="delta ' + (better ? 'better' : 'worse') + '">' + (better ? '▲' : '▼') + '</span>';
}

function render() {
  const runs = query('SELECT * FROM runs ORDER BY started_at ASC');
  const skills = query('SELECT * FROM run_skills');
  const regs = query('SELECT * FROM run_regressions');

  // Attach skills and regressions to runs
  runs.forEach(r => {
    r.skills = {};
    r.regressions = {};
    skills.filter(s => s.run_id === r.id).forEach(s => r.skills[s.skill_name] = !!s.invoked);
    regs.filter(g => g.run_id === r.id).forEach(g => r.regressions[g.name] = { occurrences: g.occurrences, max: g.max_allowed, passed: !!g.passed });
  });

  if (!runs.length) {
    document.getElementById('loading').style.display = 'none';
    document.getElementById('app').style.display = 'block';
    document.getElementById('app').innerHTML = '<div class="empty-state"><h2>No runs recorded yet</h2><p>Run <code>mage scenario:analyze</code> to record results.</p></div>';
    return;
  }

  // Group by scenario
  const scenarios = [...new Set(runs.map(r => r.scenario))];
  const byScenario = {};
  scenarios.forEach(s => byScenario[s] = runs.filter(r => r.scenario === s));

  const latest = runs[runs.length - 1];

  let html = '';

  // Summary cards
  html += '<div class="summary">';
  html += '<div class="card"><div class="card-value">' + runs.length + '</div><div class="card-label">Total Runs</div></div>';
  html += '<div class="card"><div class="card-value">' + scenarios.length + '</div><div class="card-label">Scenarios</div></div>';
  html += '<div class="card"><div class="card-label">Latest Score</div><div class="card-value ' + passClass(latest.passed) + '">' + pct(latest.score) + '%</div></div>';
  html += '</div>';

  // Tabs
  html += '<div class="tabs">';
  scenarios.forEach((s, i) => {
    html += '<div class="tab ' + (i === 0 ? 'active' : '') + '" onclick="switchTab(\'' + s + '\')">' + s + '</div>';
  });
  if (scenarios.length > 0) {
    html += '<div class="tab" onclick="switchTab(\'comparison\')">⚖️ Compare</div>';
  }
  html += '</div>';

  // Per-scenario tabs
  scenarios.forEach((name, si) => {
    const sRuns = byScenario[name];
    html += '<div class="tab-content ' + (si === 0 ? 'active' : '') + '" id="tab-' + name + '">';

    if (sRuns.length > 1) {
      html += '<div class="chart-grid">';
      ['Score', 'Duration (minutes)', 'Agent Turns', 'azd up Attempts'].forEach((title, ci) => {
        const cid = ['score','duration','turns','azdups'][ci];
        html += '<div class="chart-box"><h3>' + title + '</h3><canvas id="chart-' + cid + '-' + name + '" height="200"></canvas></div>';
      });
      html += '</div>';
    }

    // Table
    html += '<table><thead><tr>';
    ['#','Date','Session','Commit','Score','Status','Duration','Turns','azd up','Bicep','Deploy'].forEach(h => html += '<th>' + h + '</th>');
    html += '</tr></thead><tbody>';
    sRuns.slice().reverse().forEach((r, i) => {
      const idx = sRuns.length - i;
      html += '<tr onclick="toggleDetail(\'detail-' + name + '-' + i + '\')" style="cursor:pointer">';
      html += '<td>' + idx + '</td>';
      html += '<td>' + dateShort(r.started_at) + '</td>';
      html += '<td class="mono">' + shortID(r.session_id) + '</td>';
      html += '<td class="mono">' + shortCommit(r.git_commit) + '</td>';
      html += '<td><span class="score-bar"><span class="score-fill ' + scoreClass(r.score) + '" style="width:' + pct(r.score) + '%"></span></span> ' + pct(r.score) + '%</td>';
      html += '<td><span class="badge ' + passClass(r.passed) + '">' + (r.passed ? 'PASS' : 'FAIL') + '</span></td>';
      html += '<td>' + formatDur(r.duration_sec) + '</td>';
      html += '<td>' + r.total_turns + '</td>';
      html += '<td>' + r.azd_up_attempts + '</td>';
      html += '<td>' + r.bicep_edits + '</td>';
      html += '<td>' + passFail(r.deployed) + '</td>';
      html += '</tr>';

      // Detail row
      html += '<tr id="detail-' + name + '-' + i + '" style="display:none"><td colspan="11"><div class="detail-section">';
      const sk = Object.keys(r.skills);
      if (sk.length) {
        html += '<h4>Skills</h4>';
        sk.sort().forEach(k => html += '<div class="metric-row"><span class="metric-label">' + k + '</span><span>' + passFail(r.skills[k]) + '</span></div>');
      }
      const rk = Object.keys(r.regressions);
      if (rk.length) {
        html += '<h4 style="margin-top:12px">Regressions</h4>';
        rk.sort().forEach(k => {
          const g = r.regressions[k];
          html += '<div class="metric-row"><span class="metric-label">' + k + '</span><span>' + g.occurrences + '/' + g.max + ' ' + passFail(g.passed) + '</span></div>';
        });
      }
      html += '</div></td></tr>';
    });
    html += '</tbody></table></div>';
  });

  // Comparison tab
  html += '<div class="tab-content" id="tab-comparison">';
  scenarios.forEach(name => {
    const sRuns = byScenario[name];
    if (sRuns.length < 2) {
      html += '<p style="color:var(--text-dim);margin-bottom:24px">' + name + ': Need at least 2 runs for comparison.</p>';
      return;
    }
    const first = sRuns[0], last = sRuns[sRuns.length - 1];
    html += '<h2 style="margin-bottom:16px">' + name + ': Run #1 vs Latest (#' + sRuns.length + ')</h2>';
    html += '<div class="compare-grid">';

    // First run card
    html += '<div class="compare-card"><h3>🔴 Run #1 <span class="badge ' + passClass(first.passed) + '">' + (first.passed ? 'PASS' : 'FAIL') + '</span></h3>';
    html += '<div style="font-size:13px;color:var(--text-dim);margin-bottom:12px">' + dateShort(first.started_at) + ' · ' + shortID(first.session_id) + '</div>';
    html += '<div class="metric-row"><span class="metric-label">Score</span><span class="metric-value">' + pct(first.score) + '%</span></div>';
    html += '<div class="metric-row"><span class="metric-label">Duration</span><span class="metric-value">' + formatDur(first.duration_sec) + '</span></div>';
    html += '<div class="metric-row"><span class="metric-label">Turns</span><span class="metric-value">' + first.total_turns + '</span></div>';
    html += '<div class="metric-row"><span class="metric-label">azd up</span><span class="metric-value">' + first.azd_up_attempts + '</span></div>';
    html += '<div class="metric-row"><span class="metric-label">Bicep edits</span><span class="metric-value">' + first.bicep_edits + '</span></div>';
    html += '</div>';

    // Latest run card
    html += '<div class="compare-card"><h3>🟢 Run #' + sRuns.length + ' (Latest) <span class="badge ' + passClass(last.passed) + '">' + (last.passed ? 'PASS' : 'FAIL') + '</span></h3>';
    html += '<div style="font-size:13px;color:var(--text-dim);margin-bottom:12px">' + dateShort(last.started_at) + ' · ' + shortID(last.session_id) + '</div>';
    html += '<div class="metric-row"><span class="metric-label">Score</span><span class="metric-value">' + pct(last.score) + '%' + deltaHTML(last.score, first.score, false) + '</span></div>';
    html += '<div class="metric-row"><span class="metric-label">Duration</span><span class="metric-value">' + formatDur(last.duration_sec) + deltaHTML(last.duration_sec, first.duration_sec, true) + '</span></div>';
    html += '<div class="metric-row"><span class="metric-label">Turns</span><span class="metric-value">' + last.total_turns + deltaHTML(last.total_turns, first.total_turns, true) + '</span></div>';
    html += '<div class="metric-row"><span class="metric-label">azd up</span><span class="metric-value">' + last.azd_up_attempts + deltaHTML(last.azd_up_attempts, first.azd_up_attempts, true) + '</span></div>';
    html += '<div class="metric-row"><span class="metric-label">Bicep edits</span><span class="metric-value">' + last.bicep_edits + deltaHTML(last.bicep_edits, first.bicep_edits, true) + '</span></div>';
    html += '</div></div>';
  });
  html += '</div>';

  document.getElementById('loading').style.display = 'none';
  document.getElementById('app').style.display = 'block';
  document.getElementById('app').innerHTML = html;

  // Create charts
  const chartColors = { score: '#58a6ff', duration: '#bc8cff', turns: '#d29922', azdups: '#f85149' };
  const chartOpts = {
    responsive: true,
    plugins: { legend: { display: false } },
    scales: {
      x: { grid: { color: '#30363d' }, ticks: { color: '#8b949e' } },
      y: { grid: { color: '#30363d' }, ticks: { color: '#8b949e' }, beginAtZero: true }
    }
  };

  scenarios.forEach(name => {
    const sRuns = byScenario[name];
    if (sRuns.length < 2) return;
    const labels = sRuns.map(r => dateShort(r.started_at).split(' ')[0] + ' ' + dateShort(r.started_at).split(' ')[1]);

    function mk(id, data, color, label) {
      const el = document.getElementById(id);
      if (!el) return;
      new Chart(el, {
        type: 'line',
        data: { labels, datasets: [{ label, data, borderColor: color, backgroundColor: color + '22', fill: true, tension: 0.3, pointRadius: 4 }] },
        options: chartOpts
      });
    }
    mk('chart-score-' + name, sRuns.map(r => pct(r.score)), chartColors.score, 'Score %');
    mk('chart-duration-' + name, sRuns.map(r => (r.duration_sec / 60).toFixed(1)), chartColors.duration, 'Minutes');
    mk('chart-turns-' + name, sRuns.map(r => r.total_turns), chartColors.turns, 'Turns');
    mk('chart-azdups-' + name, sRuns.map(r => r.azd_up_attempts), chartColors.azdups, 'Attempts');
  });
}

function switchTab(name) {
  document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
  document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
  const el = document.getElementById('tab-' + name);
  if (el) el.classList.add('active');
  document.querySelectorAll('.tab').forEach(t => {
    if (t.textContent.trim() === name || (name === 'comparison' && t.textContent.includes('Compare')))
      t.classList.add('active');
  });
}

function toggleDetail(id) {
  const el = document.getElementById(id);
  if (el) el.style.display = el.style.display === 'none' ? 'table-row' : 'none';
}

loadDB();
</script>
</body>
</html>`

func formatDuration(sec int) string {
	if sec < 60 {
		return fmt.Sprintf("%ds", sec)
	}
	m := sec / 60
	s := sec % 60
	if m < 60 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	h := m / 60
	m = m % 60
	return fmt.Sprintf("%dh %dm", h, m)
}
