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
  instructions: string;
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
    throw new Error(errorMsg);
  }

  if (res.status === 204) {
    return {} as T;
  }

  return (await res.json()) as T;
}

export const api = {
  // Auth Token Management
  getToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem(TOKEN_KEY);
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

  // Public / Unauthenticated Invite Validation
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
    const res = await request<{ sessions?: SessionItem[]; interviews?: SessionItem[] }>(
      `/api/v1/sessions${qs}`
    );
    return res.sessions || res.interviews || [];
  },

  async getSession(id: string): Promise<SessionItem> {
    return await request<SessionItem>(`/api/v1/sessions/${id}`);
  },

  async getSessionTurns(id: string): Promise<SessionTurn[]> {
    const res = await request<{ turns: SessionTurn[] }>(`/api/v1/sessions/${id}/turns`);
    return res.turns || [];
  },

  // Evaluation Dossier
  async getSessionEvaluation(id: string): Promise<EvaluationReport> {
    return await request<EvaluationReport>(`/api/v1/sessions/${id}/evaluation`);
  },

  async enqueueEvaluation(id: string): Promise<{ status: string }> {
    return await request<{ status: string }>(`/api/v1/sessions/${id}/evaluation`, {
      method: "POST",
    });
  },
};
