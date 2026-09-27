export type AnomalyType =
  | 'REAL_ANOMALY'
  | 'EXPLAINABLE_ANOMALY'
  | 'FALSE_POSITIVE'
  | 'DATA_QUALITY'

export type Severity = 'HIGH' | 'MEDIUM' | 'LOW'

export interface User {
  id: number
  email: string
  name: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface MeterSummary {
  meter_id: string
  name: string
  location: string
  status: string
  consumption_kwh: number
}

export interface MeterDetail extends MeterSummary {
  avg_voltage_v: number
  avg_current_a: number
  avg_power_factor: number
}

export interface Reading {
  timestamp: string
  consumption_kwh: number
  voltage_v: number
  current_a: number
  power_factor: number
  status: string
}

export interface DashboardSummary {
  meters_count: number
  total_consumption_kwh: number
  anomalies_detected: number
  high_priority: number
  last_analysis: {
    status: string
    stage: string
    started_at: string
    finished_at: string | null
  } | null
}

export interface AnalysisRun {
  id: number
  status: string
  stage: string
  started_at: string
  finished_at: string | null
  meters_analyzed: number
  anomalies_found: number
  high_priority: number
}

export interface RelatedEvent {
  timestamp: string
  type: string
  description: string
}

export interface AnomalyEvidence {
  peak_zscore: number
  avg_zscore: number
  voltage_baseline_v: number
  voltage_actual_start_v: number
  voltage_actual_end_v: number
  power_factor_baseline: number
  power_factor_actual_start: number
  power_factor_actual_end: number
  data_quality_flag_count: number
  data_quality_sample_size: number
  related_event?: RelatedEvent
  ongoing: boolean
  onset_hours: number
  onset?: 'STEP' | 'GRADUAL'
}

export interface Anomaly {
  id: number
  meter_id: string
  detected_at: string
  window_start: string
  window_end: string
  type: AnomalyType
  severity: Severity
  confidence: number
  reason: string
  recommended_action: string
  status: string
  baseline_kwh: number
  actual_kwh: number
  variation_pct: number
  evidence: AnomalyEvidence
}
