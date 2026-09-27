import type { Severity } from '../api/types'

// The meter status vocabulary (OK/Alert/Critical) is separate from the raw
// anomaly severity (Medium/High) - two different vocabularies for two
// different questions (challenge brief, section 6).
export function meterStatus(severity: Severity | undefined) {
  if (severity === 'HIGH') return 'CRITICAL'
  if (severity === 'MEDIUM' || severity === 'LOW') return 'ALERT'
  return 'OK'
}
