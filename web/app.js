// web/app.js

const btnStart = document.getElementById("btn-start");
const statusEl = document.getElementById("status");
const simPill = document.getElementById("sim-pill");
const simIdEl = document.getElementById("sim-id");
const wsPill = document.getElementById("ws-pill");
const wsStatusEl = document.getElementById("ws-status");

const metricEntities = document.getElementById("metric-entities");
const metricTick = document.getElementById("metric-tick");
const metricLatency = document.getElementById("metric-latency");
const metricCompute = document.getElementById("metric-compute");

const canvas = document.getElementById("canvas");
const ctx = canvas.getContext("2d");

let currentSimId = null;
let ws = null;
let lastTickTime = null;

btnStart.addEventListener("click", async () => {
  const name = document.getElementById("sim-name").value || "Demo Simulation";
  const entities = parseInt(document.getElementById("sim-entities").value, 10) || 200;
  const tickRateMs = parseInt(document.getElementById("sim-tick").value, 10) || 50;
  const scenario = document.getElementById("sim-scenario").value || "harvest";

  try {
    setStatus("Creating simulation...", "info");

    // Create simulation via REST
    const res = await fetch("/simulations", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name,
        entities,
        tick_rate_ms: tickRateMs,
        scenario_type: scenario,
      }),
    });

    if (!res.ok) {
      const text = await res.text();
      setStatus("Failed to create simulation: " + text, "error");
      return;
    }

    const sim = await res.json();
    currentSimId = sim.id;
    simIdEl.textContent = sim.id;
    simPill.style.display = "inline-flex";

    setStatus("Starting simulation...", "info");

    // Start simulation
    const startRes = await fetch(`/simulations/${currentSimId}/start`, {
      method: "POST",
    });

    if (!startRes.ok) {
      const text = await startRes.text();
      setStatus("Failed to start simulation: " + text, "error");
      return;
    }

    setStatus("Simulation running. Connecting to WebSocket...", "info");

    // Connect WebSocket
    openWebSocket(currentSimId);
  } catch (err) {
    console.error(err);
    setStatus("Unexpected error: " + err.message, "error");
  }
});

function openWebSocket(simId) {
  if (ws) {
    ws.close();
    ws = null;
  }

  const proto = window.location.protocol === "https:" ? "wss" : "ws";
  const host = window.location.host;
  const url = `${proto}://${host}/ws/simulations/${encodeURIComponent(simId)}`;

  ws = new WebSocket(url);
  wsStatusEl.textContent = "Connecting...";
  wsPill.style.display = "inline-flex";

  ws.onopen = () => {
    wsStatusEl.textContent = "Connected";
    setStatus("Connected to WebSocket stream.", "ok");
  };

  ws.onclose = () => {
    wsStatusEl.textContent = "Closed";
    setStatus("WebSocket closed.", "warn");
  };

  ws.onerror = () => {
    wsStatusEl.textContent = "Error";
    setStatus("WebSocket error.", "error");
  };

  ws.onmessage = (event) => {
    const now = performance.now();
    let update;
    try {
      update = JSON.parse(event.data);
    } catch (e) {
      console.error("Failed to parse WS message", e);
      return;
    }

    // update shape:
    // {
    //   simulation_id, tick, entities: [{id,x,y,vx,vy,battery,status}],
    //   avg_compute_ms, worker_count, completed_at
    // }

    renderFrame(update);

    metricEntities.textContent = `Entities: ${update.entities?.length ?? 0}`;
    metricTick.textContent = `Tick: ${update.tick}`;
    metricCompute.textContent = `Avg compute: ${update.avg_compute_ms?.toFixed(2) ?? "--"} ms`;

    if (lastTickTime != null) {
      const dt = now - lastTickTime;
      metricLatency.textContent = `Latency (client tick): ${dt.toFixed(1)} ms`;
    } else {
      metricLatency.textContent = "Latency: measuring…";
    }
    lastTickTime = now;
  };
}

function setStatus(msg, type) {
  statusEl.textContent = msg;
  statusEl.style.color = {
    ok: "#a7f3d0",
    info: "#93c5fd",
    warn: "#facc15",
    error: "#fecaca",
  }[type] || "#9ca3af";
}

function renderFrame(update) {
  if (!update || !Array.isArray(update.entities)) {
    return;
  }

  const entities = update.entities;
  const w = canvas.width;
  const h = canvas.height;

  // Clear canvas with slightly transparent fill for a subtle trail effect.
  ctx.fillStyle = "rgba(4,6,20,0.9)";
  ctx.fillRect(0, 0, w, h);

  // Draw a faint grid for visual reference.
  drawGrid(ctx, w, h);

  for (const e of entities) {
    const x = (e.x / 100) * w;
    const y = (e.y / 100) * h;

    const battery = e.battery ?? 100;
    let color;
    if (battery <= 0) {
      color = "#4b5563"; // dead
    } else if (battery < 20) {
      color = "#f97373"; // low
    } else {
      color = "#22c55e"; // normal
    }

    drawEntity(ctx, x, y, color);
  }
}

function drawEntity(ctx, x, y, color) {
  const r = 6;
  ctx.beginPath();
  ctx.arc(x, y, r, 0, Math.PI * 2);
  ctx.fillStyle = color;
  ctx.fill();

  // tiny glow
  const grd = ctx.createRadialGradient(x, y, 0, x, y, 16);
  grd.addColorStop(0, "rgba(34,197,94,0.55)");
  grd.addColorStop(1, "rgba(15,23,42,0)");
  ctx.fillStyle = grd;
  ctx.beginPath();
  ctx.arc(x, y, 16, 0, Math.PI * 2);
  ctx.fill();
}

function drawGrid(ctx, w, h) {
  ctx.save();
  ctx.strokeStyle = "rgba(148,163,184,0.12)";
  ctx.lineWidth = 1;

  const stepX = w / 10;
  const stepY = h / 10;

  ctx.beginPath();
  for (let x = 0; x <= w; x += stepX) {
    ctx.moveTo(x, 0);
    ctx.lineTo(x, h);
  }
  for (let y = 0; y <= h; y += stepY) {
    ctx.moveTo(0, y);
    ctx.lineTo(w, y);
  }
  ctx.stroke();
  ctx.restore();
}

// Initial clear
renderFrame({ entities: [] });
