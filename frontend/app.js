const $ = (selector) => document.querySelector(selector);
const state = { token: localStorage.getItem("hm_token"), jobs: [], selectedJob: null, poller: null };

function setMessage(text, error = false) {
  const target = $("#dashboard-message");
  target.textContent = text || "";
  target.className = error ? "error" : "message";
}

async function api(path, options = {}) {
  const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
  if (state.token) headers.Authorization = `Bearer ${state.token}`;
  const response = await fetch(`/api/v1${path}`, { ...options, headers });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(body.error?.message || body.message || `Request failed (${response.status})`);
  return body;
}

function formData(form) { return Object.fromEntries(new FormData(form).entries()); }
function escapeHtml(value) { const div = document.createElement("div"); div.textContent = value ?? ""; return div.innerHTML; }

function showDashboard() {
  $("#auth-view").classList.add("hidden");
  $("#candidate-view").classList.add("hidden");
  $("#dashboard-view").classList.remove("hidden");
  loadDashboard();
  clearInterval(state.poller);
  state.poller = setInterval(refreshLiveData, 5000);
}

async function loadDashboard() {
  try { await Promise.all([loadJobs(), loadInterviews()]); } catch (err) { setMessage(err.message, true); }
}

async function refreshLiveData() {
  if (!state.token) return;
  try {
    await loadInterviews();
    if (state.selectedJob) await loadCandidates(state.selectedJob.id);
  } catch (_) { /* keep the last visible data while a service is restarting */ }
}

async function loadJobs() {
  const result = await api("/jobs?limit=100");
  state.jobs = result.jobs || [];
  const container = $("#jobs");
  container.innerHTML = state.jobs.length ? state.jobs.map((job) => `
    <div class="list-item"><div><strong>${escapeHtml(job.title)}</strong><small>${escapeHtml(job.description || "No description")}</small></div>
    <button class="secondary" data-job-id="${job.id}">Open</button></div>`).join("") : "<p class='muted'>Create your first job to begin.</p>";
  container.querySelectorAll("[data-job-id]").forEach((button) => button.addEventListener("click", () => selectJob(button.dataset.jobId)));
}

async function selectJob(jobID) {
  state.selectedJob = state.jobs.find((job) => job.id === jobID);
  if (!state.selectedJob) return;
  $("#job-workspace").classList.remove("hidden");
  $("#selected-job-title").textContent = state.selectedJob.title;
  const savedDocket = JSON.parse(localStorage.getItem(`docket:${jobID}`) || "null");
  $("#invite-docket-id").value = savedDocket?.id || "";
  $("#selected-docket").textContent = savedDocket ? `Docket: ${savedDocket.title}` : "No docket selected";
  await loadCandidates(jobID);
}

async function loadCandidates(jobID) {
  const result = await api(`/jobs/${jobID}/candidates`);
  const candidates = result.candidates || [];
  $("#candidates").innerHTML = candidates.length ? candidates.map((candidate) => `
    <div class="list-item"><div><strong>${escapeHtml(candidate.full_name)}</strong><small>${escapeHtml(candidate.email)} · ${escapeHtml(candidate.interview_status || "not invited")}</small></div>
    <button class="secondary" data-invite-candidate="${candidate.id}">Generate invite</button></div>`).join("") : "<p class='muted'>No candidates yet.</p>";
  $("#candidates").querySelectorAll("[data-invite-candidate]").forEach((button) => button.addEventListener("click", () => createInvite(button.dataset.inviteCandidate)));
}

async function loadInterviews() {
  const result = await api("/interviews?limit=100");
  const interviews = result.interviews || [];
  $("#interviews").innerHTML = interviews.length ? interviews.map((interview) => `
    <div class="list-item"><div><strong>${escapeHtml(interview.candidate.full_name)}</strong><small>${escapeHtml(interview.docket.title)} · ${escapeHtml(interview.status)}${interview.feedback ? ` · ${interview.feedback.overall_score}/10` : ""}</small></div>
    <button class="secondary" data-interview-id="${interview.id}">Open</button></div>`).join("") : "<p class='muted'>Invites will appear here once generated.</p>";
  $("#interviews").querySelectorAll("[data-interview-id]").forEach((button) => button.addEventListener("click", () => openInterview(button.dataset.interviewId)));
}

async function openInterview(interviewID) {
  try {
    const detail = await api(`/interviews/${interviewID}`);
    const panel = $("#interview-detail");
    panel.classList.remove("hidden");
    panel.innerHTML = `
      <div style="display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:12px;">
        <div>
          <h2>${escapeHtml(detail.candidate.full_name)} — ${escapeHtml(detail.docket.title)}</h2>
          <span class="pill pill-indigo" style="margin-top:6px;">Status: ${escapeHtml(detail.status)}</span>
        </div>
        <button class="btn btn-outline small" onclick="document.querySelector('#interview-detail').classList.add('hidden')">Close ✕</button>
      </div>
      <p class="muted small">This section refreshes automatically as feedback evaluation completes.</p>
    `;
    if (detail.status !== "feedback_ready") return;
    const feedback = await api(`/interviews/${interviewID}/feedback`);
    panel.innerHTML += `
      <div class="feedback-hero">
        <div class="score-badge">${feedback.overall_score}/10</div>
        <div class="rec-badge rec-${escapeHtml(feedback.recommendation?.toLowerCase().replace(/[^a-z0-9_]/g, ''))}">Recommendation: ${escapeHtml(feedback.recommendation)}</div>
      </div>
      <p style="margin: 14px 0; font-size: 0.95rem; line-height:1.6; color: var(--slate-700);">${escapeHtml(feedback.summary)}</p>
      
      <h3 style="margin-top:24px; margin-bottom:12px;">Rubric Evaluation Breakdown</h3>
      ${feedback.scores.map((score) => `
        <div class="feedback-score">
          <div style="display:flex; justify-content:space-between; font-weight:700; margin-bottom:4px;">
            <span>${escapeHtml(score.criterion_name)}</span>
            <span style="color:var(--primary);">${score.score} / ${score.max_score}</span>
          </div>
          <small style="display:block; color:var(--slate-600); line-height:1.45;">${escapeHtml(score.evidence)}</small>
        </div>
      `).join("")}
      
      <div class="transcript">
        <h3 style="margin-bottom:10px;">Full Interview Transcript</h3>
        ${feedback.transcript.map((turn) => `
          <p><strong>${escapeHtml(turn.role === 'assistant' ? 'Alex (AI Interviewer)' : detail.candidate.full_name)}:</strong> ${escapeHtml(turn.content)}</p>
        `).join("")}
      </div>
    `;
  } catch (err) { setMessage(err.message, true); }
}

async function createInvite(candidateID) {
  const docketID = $("#invite-docket-id").value.trim();
  const days = Number($("#invite-days").value || 7);
  if (!docketID) { setMessage("Create a docket first, or paste its ID into the invite field.", true); return; }
  try {
    const invite = await api(`/candidates/${candidateID}/invite`, { method: "POST", body: JSON.stringify({ docket_id: docketID, expires_in_days: days }) });
    const link = `${location.origin}/interview/${invite.token}`;
    $("#invite-result").innerHTML = `<strong>Invite ready</strong><br/><a href="${link}" target="_blank" rel="noopener">${link}</a><br/><button class="secondary" id="copy-invite">Copy link</button>`;
    $("#copy-invite").onclick = async () => { await navigator.clipboard.writeText(link); setMessage("Invite link copied."); };
    await Promise.all([loadCandidates(state.selectedJob.id), loadInterviews()]);
  } catch (err) { setMessage(err.message, true); }
}

function setupDashboard() {
  $("#job-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const data = formData(form);
    try {
      await api("/jobs", {
        method: "POST",
        body: JSON.stringify({
          title: data.title,
          description: data.description,
          required_skills: data.skills.split(",").map((s) => s.trim()).filter(Boolean)
        })
      });
      if (form) form.reset();
      setMessage("Job created.");
      await loadJobs();
    } catch (err) {
      setMessage(err.message, true);
    }
  });

  $("#docket-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedJob) return;
    const form = event.currentTarget;
    const data = formData(form);
    try {
      const docket = await api(`/jobs/${state.selectedJob.id}/dockets`, {
        method: "POST",
        body: JSON.stringify({
          title: data.title,
          instructions: data.instructions,
          duration_mins: Number(data.duration_mins),
          questions: [{ question_text: data.question, question_type: "technical", order_index: 1 }],
          rubric: [{ name: data.rubric_name, description: data.rubric_description, max_score: Number(data.max_score), order_index: 1 }]
        })
      });
      localStorage.setItem(`docket:${state.selectedJob.id}`, JSON.stringify({ id: docket.id, title: docket.title }));
      $("#invite-docket-id").value = docket.id;
      $("#selected-docket").textContent = `Docket: ${docket.title}`;
      if (form) form.reset();
      setMessage("Docket created. You can now generate invites.");
    } catch (err) {
      setMessage(err.message, true);
    }
  });

  $("#candidate-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedJob) return;
    const form = event.currentTarget;
    const data = formData(form);
    try {
      await api(`/jobs/${state.selectedJob.id}/candidates`, {
        method: "POST",
        body: JSON.stringify(data)
      });
      if (form) form.reset();
      setMessage("Candidate added.");
      await loadCandidates(state.selectedJob.id);
    } catch (err) {
      setMessage(err.message, true);
    }
  });

  $("#refresh").onclick = loadDashboard;
  $("#logout").onclick = () => {
    localStorage.removeItem("hm_token");
    state.token = null;
    clearInterval(state.poller);
    $("#dashboard-view").classList.add("hidden");
    $("#auth-view").classList.remove("hidden");
  };
}

function setupAuth() {
  document.querySelectorAll("[data-auth]").forEach((tab) =>
    tab.addEventListener("click", () => {
      document.querySelectorAll("[data-auth]").forEach((button) => button.classList.toggle("active", button === tab));
      $("#login-form").classList.toggle("hidden", tab.dataset.auth !== "login");
      $("#register-form").classList.toggle("hidden", tab.dataset.auth !== "register");
    })
  );

  ["login", "register"].forEach((kind) =>
    $("#" + kind + "-form").addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = event.currentTarget;
      const data = formData(form);
      try {
        const result = await api(`/auth/${kind}`, {
          method: "POST",
          body: JSON.stringify(data)
        });
        state.token = result.token;
        localStorage.setItem("hm_token", result.token);
        if (form) form.reset();
        showDashboard();
      } catch (err) {
        $("#auth-error").textContent = err.message;
      }
    })
  );
}

function base64FromBlob(blob) { return new Promise((resolve, reject) => { const reader = new FileReader(); reader.onload = () => resolve(reader.result.split(",")[1]); reader.onerror = reject; reader.readAsDataURL(blob); }); }
function bytesFromBase64(value) { const raw = atob(value); return Uint8Array.from(raw, (char) => char.charCodeAt(0)); }
function stripMarkdown(text) {
  return (text || "")
    .replace(/\*\*([^*]+)\*\*/g, "$1")   // **bold**
    .replace(/\*([^*]+)\*/g, "$1")         // *italic*
    .replace(/__([^_]+)__/g, "$1")          // __bold__
    .replace(/_([^_]+)_/g, "$1")           // _italic_
    .replace(/~~([^~]+)~~/g, "$1")          // ~~strikethrough~~
    .replace(/`{1,3}[^`]*`{1,3}/g, "")     // `code` / ```block```
    .replace(/^#{1,6}\s+/gm, "")           // # Headings
    .replace(/^>\s+/gm, "")                // > blockquote
    .replace(/!?\[([^\]]+)\]\([^)]+\)/g, "$1") // [links](url)
    .replace(/[-*+]\s+/g, "")              // list bullets
    .replace(/\n{2,}/g, ". ")              // double newlines → pause
    .replace(/\n/g, " ")                   // remaining newlines → space
    .trim();
}

async function showCandidate(token) {
  $("#auth-view").classList.add("hidden"); $("#dashboard-view").classList.add("hidden"); $("#candidate-view").classList.remove("hidden");
  const info = $("#candidate-info"), stateLabel = $("#candidate-state"), error = $("#candidate-error"), connect = $("#connect"), record = $("#record"), complete = $("#complete"), conversation = $("#conversation");
  let socket, recorder, chunks = [], currentAI, audioChunks = [], completeSent = false, heartbeat;
  const setState = (text) => {
    stateLabel.textContent = text;
    stateLabel.className = "stage-status-badge";
    const lower = (text || "").toLowerCase();
    if (lower.includes("listening")) stateLabel.classList.add("status-listening");
    else if (lower.includes("speaking")) stateLabel.classList.add("status-speaking");
    else if (lower.includes("connected") || lower.includes("ready")) stateLabel.classList.add("status-ready");
    else if (lower.includes("complete")) stateLabel.classList.add("status-completed");
  };
  const addBubble = (role, text = "") => { const node = document.createElement("div"); node.className = `bubble ${role}`; node.textContent = text; conversation.append(node); conversation.scrollTop = conversation.scrollHeight; return node; };
  try {
    const invite = await fetch(`/api/v1/invites/validate/${encodeURIComponent(token)}`).then(async (response) => { const body = await response.json(); if (!response.ok || !body.valid) throw new Error(body.error?.message || "This invite is invalid or expired."); return body; });
    info.textContent = `${invite.candidate.full_name}, you are interviewing for ${invite.job_role.title}. Docket: ${invite.docket_title}.`;
    connect.disabled = false; setState("Ready to connect");
  } catch (err) {
    const msg = err.message || "";
    if (msg.toLowerCase().includes("already completed") || msg.toLowerCase().includes("completed")) {
      setState("Interview completed");
      info.textContent = "This interview has already been completed. Thank you!";
    } else {
      error.textContent = msg;
      setState("Invite unavailable");
    }
    return;
  }
  function speakFallback(text) { if ("speechSynthesis" in window && text) { speechSynthesis.cancel(); const utt = new SpeechSynthesisUtterance(stripMarkdown(text)); utt.rate = 1; utt.pitch = 1; speechSynthesis.speak(utt); } }
  async function finishRecording() {
    const blob = new Blob(chunks, { type: recorder.mimeType || "audio/webm" }); chunks = [];
    if (!blob.size) { error.textContent = "No microphone audio was captured."; return; }
    setState("Processing your answer"); record.disabled = true;
    socket.send(JSON.stringify({ type: "audio_end", data: await base64FromBlob(blob), mime_type: blob.type }));
  }
  async function toggleRecording() {
    error.textContent = "";
    if (recorder?.state === "recording") { recorder.stop(); record.textContent = "Start speaking"; return; }
    try { const stream = await navigator.mediaDevices.getUserMedia({ audio: true }); recorder = new MediaRecorder(stream); chunks = []; recorder.ondataavailable = (event) => chunks.push(event.data); recorder.onstop = async () => { stream.getTracks().forEach((track) => track.stop()); await finishRecording(); }; recorder.start(); record.textContent = "Stop & send answer"; setState("Listening"); } catch (err) { error.textContent = "Microphone access is required to answer: " + err.message; }
  }
  function connectInterview() {
    if (socket?.readyState === WebSocket.OPEN) return;
    setState("Connecting"); error.textContent = ""; connect.disabled = true;
    const protocol = location.protocol === "https:" ? "wss:" : "ws:";
    socket = new WebSocket(`${protocol}//${location.host}/ws/interview/?token=${encodeURIComponent(token)}`);
    socket.onopen = () => { heartbeat = setInterval(() => socket.readyState === WebSocket.OPEN && socket.send(JSON.stringify({ type: "heartbeat" })), 10000); };
    socket.onmessage = async (event) => {
      const message = JSON.parse(event.data);
      if (message.type === "connected") { setState("Connected — ready to listen"); record.disabled = false; complete.disabled = false; return; }
      if (message.type === "ai_text_chunk") { if (!currentAI) currentAI = addBubble("assistant"); if (message.text) currentAI.textContent += message.text; return; }
      if (message.type === "ai_text_end") { setState("AI speaking"); return; }
      if (message.type === "ai_audio_chunk") { audioChunks.push(bytesFromBase64(message.data)); return; }
      if (message.type === "ai_audio_end") { const text = currentAI?.textContent || ""; if (audioChunks.length) { const audio = new Audio(URL.createObjectURL(new Blob(audioChunks, { type: "audio/mpeg" }))); audio.play().catch(() => speakFallback(text)); } else speakFallback(text); audioChunks = []; currentAI = null; setState("Ready to listen"); record.disabled = false; return; }
      if (message.type === "interview_complete") { completeSent = true; clearInterval(heartbeat); setState("Interview complete — feedback is being prepared"); record.disabled = true; complete.disabled = true; return; }
      if (message.type === "error") { error.textContent = message.message || message.code || "Interview error"; setState("Connection needs attention"); }
    };
    socket.onclose = () => { clearInterval(heartbeat); if (!completeSent) { setState("Disconnected — select Connect to resume"); connect.disabled = false; record.disabled = true; } };
    socket.onerror = () => { error.textContent = "Unable to connect to the interview service."; };
  }
  connect.onclick = connectInterview; record.onclick = toggleRecording; complete.onclick = () => { if (socket?.readyState === WebSocket.OPEN) { setState("Completing interview"); socket.send(JSON.stringify({ type: "end_interview" })); } };
}

function init() {
  const token = location.pathname.match(/^\/interview\/([^/]+)$/)?.[1];
  if (token) { showCandidate(token); return; }
  setupAuth(); setupDashboard(); if (state.token) showDashboard();
}
init();
