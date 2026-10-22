/**
 * VoxAssess Unified API Client
 * Connects the frontend to Go microservices (job-candidate-service, interview-orchestrator).
 * All data is real: zero mock fixtures.
 */

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8001";
export const WS_BASE_URL =
  process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8002";

// LocalStorage Keys
const TOKEN_KEY = "vox_auth_token";
const USER_KEY = "vox_auth_user";
const ORG_KEY = "vox_auth_org";

export interface User {
  id: string;
  email: string;
  full_name: string;
  role: string;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
}

export interface AuthResponse {
  token: string;
  user: User;
  org: Organization;
  user_id?: string;
  org_id?: string;
}

export interface RawAuthResponse {
  token: string;
  user_id?: string;
  org_id?: string;
  user?: User;
  org?: Organization;
}

export interface Track {
  id: string;
  title: string;
  description: string;
  required_skills: string[];
  status?: "active" | "archived" | "calibrating" | string;
  created_at?: string;
  updated_at?: string;
}

export interface QuestionPrompt {
  id?: string;
  question_text: string;
  question_type?: string;
  order_index: number;
  follow_up_hint?: string;
}

export interface EvaluationCriterion {
  id?: string;
  name: string;
  description: string;
  max_score: number;
  order_index: number;
}

export interface Blueprint {
  id: string;
  job_id?: string;
  track_id?: string;
  title: string;
  instructions?: string;
  duration_mins: number;
  questions?: QuestionPrompt[];
  prompts?: QuestionPrompt[];
  rubric?: EvaluationCriterion[];
  criteria?: EvaluationCriterion[];
  created_at?: string;
}

export interface Participant {
  id: string;
  job_id?: string;
  track_id?: string;
  full_name: string;
  email: string;
  interview_status?: "invited" | "in_progress" | "completed" | string;
  created_at?: string;
}

export interface InviteResponse {
  invite_id: string;
  invite_link: string;
  token: string;
  expires_at: string;
}

export interface InviteValidation {
  valid: boolean;
  interview_id?: string;
  interview_status?: string;
  candidate?: {
    id: string;
    full_name: string;
    email?: string;
  };
  participant?: {
    id: string;
    full_name: string;
    email?: string;
  };
  job_role?: {
    id?: string;
    title: string;
  };
  track?: {
    id: string;
    title: string;
  };
  docket_title?: string;
  blueprint_title?: string;
  duration_mins?: number;
  expires_at?: string;
}

export interface PublicTrackResponse {
  track: Track;
  blueprint_id?: string;
  blueprint_title?: string;
  duration_mins?: number;
}

export interface SessionTurn {
  id?: string;
  turn_number: number;
  role: "agent" | "participant" | string;
  content: string;
  created_at: string;
}

export interface CriterionResult {
  criterion_id: string;
  name?: string;
  score: number;
  max_score: number;
  feedback: string;
  evidence_quote?: string;
  grounded?: boolean;
}

export interface EvaluationReport {
  id: string;
  session_id: string;
  participant_name: string;
  overall_score: number;
  recommendation: "strong_hire" | "hire" | "revisit" | "no_hire" | string;
  summary: string;
  criteria_results: CriterionResult[];
  created_at?: string;
}

export interface SessionItem {
  id: string;
  participant_id: string;
  participant_name?: string;
  participant_email?: string;
  track_id?: string;
  track_title?: string;
  blueprint_id?: string;
  blueprint_title?: string;
  status: "invited" | "in_progress" | "completed" | "feedback_ready" | string;
  turn_count?: number;
  score?: number;
  recommendation?: string;
  started_at?: string;
  completed_at?: string;
  created_at?: string;
}

// Helper: HTTP Request with automatic JWT Bearer injection
async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = typeof window !== "undefined" ? localStorage.getItem(TOKEN_KEY) : null;
  const headers = new Headers(options.headers || {});

  headers.set("Content-Type", "application/json");
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const url = `${API_BASE_URL}${path}`;
  const res = await fetch(url, {
    ...options,
    headers,
  });

  if (!res.ok) {
    let errorMsg = `HTTP ${res.status} ${res.statusText}`;
    try {
      const errJson = await res.json();
      if (typeof errJson.error === "string") errorMsg = errJson.error;
      else if (errJson.error?.message) errorMsg = errJson.error.message;
      else if (errJson.message) errorMsg = errJson.message;
    } catch {
      // Body not JSON
    }
    const error: any = new Error(errorMsg);
    error.status = res.status;
    throw error;
  }

  if (res.status === 204) {
    return {} as T;
  }

  return (await res.json()) as T;
}

function normalizeSessionItem(item: any): SessionItem {
  if (!item) return item;
  return {
    id: item.id || "",
    participant_id: item.candidate?.id || item.participant_id || "",
    participant_name: item.candidate?.full_name || item.participant_name || "Candidate",
    participant_email: item.candidate?.email || item.participant_email || "",
    track_id: item.docket?.job_role_id || item.track_id || "",
    track_title: item.docket?.title || item.track_title || "Technical Track",
    blueprint_id: item.docket?.id || item.blueprint_id || "",
    blueprint_title: item.docket?.title || item.blueprint_title || "Technical Assessment Blueprint",
    status: item.status || "invited",
    turn_count: item.turn_count || 0,
    score: item.feedback?.overall_score ?? item.score,
    recommendation: item.feedback?.recommendation || item.recommendation,
    started_at: item.started_at,
    completed_at: item.completed_at,
    created_at: item.created_at,
  };
}

function normalizeEvaluationReport(raw: any, sessionId: string): EvaluationReport {
  const scores = raw.scores || raw.criteria_results || [];
  const criteriaResults: CriterionResult[] = scores.map((s: any, idx: number) => {
    const evidence = s.evidence_quote || s.evidence || undefined;
    const rawFeedback = s.feedback || "";
    // Only keep feedback if it is distinct from the raw verbatim evidence quote
    const feedback = rawFeedback && rawFeedback !== evidence ? rawFeedback : "";

    return {
      criterion_id: s.criterion_id || s.id || `crit-${idx + 1}`,
      name: s.criterion_name || s.name || `Criterion #${idx + 1}`,
      score: s.score ?? 0,
      max_score: s.max_score ?? 10,
      feedback: feedback,
      evidence_quote: evidence,
      grounded: s.is_grounded ?? s.grounded ?? true,
    };
  });

  return {
    id: raw.id || "",
    session_id: sessionId,
    participant_name: raw.participant_name || raw.candidate?.full_name || "",
    overall_score: raw.overall_score ?? 0,
    recommendation: raw.recommendation || "revisit",
    summary: raw.summary || "",
    criteria_results: criteriaResults,
    created_at: raw.created_at,
  };
}

export const api = {
  // Auth Token Management
  getToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem(TOKEN_KEY);
  },
  isAuthenticated(): boolean {
    if (typeof window === "undefined") return false;
    return !!localStorage.getItem(TOKEN_KEY);
  },
  getUser(): User | null {
    if (typeof window === "undefined") return null;
    const str = localStorage.getItem(USER_KEY);
    return str ? JSON.parse(str) : null;
  },
  getOrg(): Organization | null {
    if (typeof window === "undefined") return null;
    const str = localStorage.getItem(ORG_KEY);
    return str ? JSON.parse(str) : null;
  },
  saveAuth(auth: RawAuthResponse, fallbackEmail?: string) {
    if (typeof window === "undefined") return;
    if (auth.token) localStorage.setItem(TOKEN_KEY, auth.token);

    let email = auth.user?.email || fallbackEmail || "";
    let orgId = auth.org_id || auth.org?.id || "";
    let userId = auth.user_id || auth.user?.id || "";

    if ((!email || !orgId) && auth.token) {
      try {
        const parts = auth.token.split(".");
        if (parts[1]) {
          const payload = JSON.parse(atob(parts[1]));
          if (payload.email) email = payload.email;
          if (payload.org_id) orgId = payload.org_id;
          if (payload.sub) userId = payload.sub;
        }
      } catch {}
    }

    const user: User = auth.user || {
      id: userId,
      email: email,
      full_name: email.split("@")[0] || "Hiring Lead",
      role: "admin",
    };
    const org: Organization = auth.org || {
      id: orgId,
      name: "Tenant Organization",
      slug: "tenant",
    };

    localStorage.setItem(USER_KEY, JSON.stringify(user));
    localStorage.setItem(ORG_KEY, JSON.stringify(org));
  },
  clearAuth() {
    if (typeof window === "undefined") return;
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    localStorage.removeItem(ORG_KEY);
  },

  // Auth Endpoints
  async login(payload: { email: string; password: string }): Promise<AuthResponse> {
    const res = await request<RawAuthResponse>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(payload),
    });
    this.saveAuth(res, payload.email);
    return {
      token: res.token,
      user: this.getUser() || { id: res.user_id || "", email: payload.email, full_name: "Lead", role: "admin" },
      org: this.getOrg() || { id: res.org_id || "", name: "Tenant", slug: "tenant" },
      user_id: res.user_id,
      org_id: res.org_id,
    };
  },

  async register(payload: {
    email: string;
    password: string;
    full_name: string;
    org_name: string;
    org_slug: string;
  }): Promise<AuthResponse> {
    const res = await request<RawAuthResponse>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify(payload),
    });
    this.saveAuth(res, payload.email);
    return {
      token: res.token,
      user: this.getUser() || { id: res.user_id || "", email: payload.email, full_name: payload.full_name, role: "admin" },
      org: this.getOrg() || { id: res.org_id || "", name: payload.org_name, slug: payload.org_slug },
      user_id: res.user_id,
      org_id: res.org_id,
    };
  },

  // Tracks (Roles)
  async getTracks(): Promise<Track[]> {
    const res = await request<{ tracks?: Track[]; jobs?: Track[] }>("/api/v1/tracks");
    return res.tracks || res.jobs || [];
  },

  async getTrack(id: string): Promise<Track> {
    return await request<Track>(`/api/v1/tracks/${id}`);
  },

  async createTrack(payload: {
    title: string;
    description: string;
    required_skills: string[];
  }): Promise<Track> {
    return await request<Track>("/api/v1/tracks", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  async deleteTrack(id: string): Promise<void> {
    await request<void>(`/api/v1/tracks/${id}`, {
      method: "DELETE",
    });
  },

  // Blueprints
  async getBlueprints(trackId: string): Promise<Blueprint[]> {
    const res = await request<{ blueprints?: Blueprint[]; dockets?: Blueprint[] }>(
      `/api/v1/tracks/${trackId}/blueprints`
    );
    return res.blueprints || res.dockets || [];
  },

  async getBlueprint(id: string): Promise<Blueprint> {
    return await request<Blueprint>(`/api/v1/blueprints/${id}`);
  },

  async createBlueprint(
    trackId: string,
    payload: {
      title: string;
      instructions: string;
      duration_mins: number;
      questions: QuestionPrompt[];
      rubric: EvaluationCriterion[];
    }
  ): Promise<Blueprint> {
    const questions = payload.questions.map((q, idx) => ({
      question_text: q.question_text,
      question_type: q.question_type || "technical",
      order_index: q.order_index || idx + 1,
      follow_up_hint: q.follow_up_hint || undefined,
    }));
    const rubric = payload.rubric.map((r, idx) => ({
      name: r.name,
      description: r.description || "Demonstrates core technical competence in this domain.",
      max_score: Number(r.max_score) || 10,
      order_index: r.order_index || idx + 1,
    }));

    return await request<Blueprint>(`/api/v1/tracks/${trackId}/blueprints`, {
      method: "POST",
      body: JSON.stringify({
        title: payload.title,
        instructions: payload.instructions,
        duration_mins: Number(payload.duration_mins) || 30,
        questions,
        rubric,
      }),
    });
  },

  // Participants & Invites
  async getParticipants(trackId: string): Promise<Participant[]> {
    const res = await request<{ participants?: Participant[]; candidates?: Participant[] }>(
      `/api/v1/tracks/${trackId}/participants`
    );
    return res.participants || res.candidates || [];
  },

  async createParticipant(
    trackId: string,
    payload: { full_name: string; email: string }
  ): Promise<Participant> {
    return await request<Participant>(`/api/v1/tracks/${trackId}/participants`, {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  async createInvite(
    participantId: string,
    payload: { docket_id: string; expires_in_days?: number }
  ): Promise<InviteResponse> {
    return await request<InviteResponse>(
      `/api/v1/participants/${participantId}/invite`,
      {
        method: "POST",
        body: JSON.stringify({
          docket_id: payload.docket_id,
          expires_in_days: payload.expires_in_days || 7,
        }),
      }
    );
  },

  // Public / Unauthenticated Candidate Application & Validation
  async getPublicTrack(trackId: string): Promise<PublicTrackResponse> {
    return await request<PublicTrackResponse>(`/api/v1/public/tracks/${trackId}`);
  },

  async publicApply(
    trackId: string,
    payload: { full_name: string; email: string }
  ): Promise<InviteResponse> {
    return await request<InviteResponse>(`/api/v1/public/tracks/${trackId}/apply`, {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  async validateInvite(token: string): Promise<InviteValidation> {
    return await request<InviteValidation>(`/api/v1/invites/validate/${encodeURIComponent(token)}`);
  },

  // Sessions & Turns
  async createSession(inviteId: string): Promise<{ id: string }> {
    return await request<{ id: string }>("/api/v1/sessions", {
      method: "POST",
      body: JSON.stringify({ invite_id: inviteId }),
    });
  },

  async getSessions(params?: { status?: string; page?: number; limit?: number }): Promise<SessionItem[]> {
    const searchParams = new URLSearchParams();
    if (params?.status) searchParams.set("status", params.status);
    if (params?.page) searchParams.set("page", String(params.page));
    if (params?.limit) searchParams.set("limit", String(params.limit));

    const qs = searchParams.toString() ? `?${searchParams.toString()}` : "";
    const res = await request<{ sessions?: any[]; interviews?: any[] }>(
      `/api/v1/sessions${qs}`
    );
    const list = res.sessions || res.interviews || [];
    return list.map(normalizeSessionItem);
  },

  async getSession(id: string): Promise<SessionItem> {
    const item = await request<any>(`/api/v1/sessions/${id}`);
    return normalizeSessionItem(item);
  },

  async getSessionTurns(id: string): Promise<SessionTurn[]> {
    const res = await request<{ turns: SessionTurn[] }>(`/api/v1/sessions/${id}/turns`);
    return res.turns || [];
  },

  // Evaluation Dossier
  async getSessionEvaluation(id: string): Promise<EvaluationReport | null> {
    try {
      const raw = await request<any>(`/api/v1/sessions/${id}/evaluation`);
      return normalizeEvaluationReport(raw, id);
    } catch (err: any) {
      if (err?.status === 404 || (typeof err?.message === "string" && err.message.includes("404"))) {
        return null;
      }
      throw err;
    }
  },

  async enqueueEvaluation(id: string): Promise<{ status: string }> {
    return await request<{ status: string }>(`/api/v1/sessions/${id}/evaluation`, {
      method: "POST",
    });
  },
};
