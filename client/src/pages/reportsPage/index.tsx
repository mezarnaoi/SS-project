import React, { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { apiFetch } from '../../utils/api';

interface ReportSummary {
  total_checks: number;
  checks_last_month: number;
  checks_this_month: number;
  expiring_next_month: ExpirationAlert[];
  fit_count: number;
  conditionally_fit_count: number;
  temporarily_unfit_count: number;
  unfit_count: number;
  ocr_success_rate: number;
  needs_review_count: number;
  avg_processing_time_ms: number;
}

interface ExpirationAlert {
  last_name: string;
  first_name: string;
  company: string;
  workplace: string;
  next_exam_date: string;
  days_until_expiry: number;
  medical_opinion: string;
}

interface AnonymizedRecord {
  control_type: string;
  medical_opinion: string;
  company: string;
  workplace: string;
  job_title: string;
  exam_date: string;
  next_exam_date: string;
  ocr_confidence: number;
}

interface PerformanceMetrics {
  total_processed: number;
  ocr_success_count: number;
  ocr_success_rate: number;
  needs_review_count: number;
  reviewed_count: number;
  avg_ocr_confidence: number;
  medical_cert_count: number;
  avg_processing_time_ms: number;
  min_processing_time_ms: number;
  max_processing_time_ms: number;
}

type Tab = 'summary' | 'expiring' | 'anonymized' | 'performance';

const fmtDate = (iso: string) => {
  if (!iso) return '—';
  const d = new Date(iso);
  if (isNaN(d.getTime()) || d.getFullYear() < 2000) return '—';
  return d.toLocaleDateString('ro-RO');
};

const urgencyClass = (days: number) => {
  if (days <= 7) return 'bg-red-50 text-red-700 border-red-200';
  if (days <= 30) return 'bg-orange-50 text-orange-700 border-orange-200';
  return 'bg-yellow-50 text-yellow-700 border-yellow-200';
};

const urgencyBadge = (days: number) => {
  if (days <= 7) return 'bg-red-100 text-red-700';
  if (days <= 30) return 'bg-orange-100 text-orange-700';
  return 'bg-yellow-100 text-yellow-700';
};

const MetricCard: React.FC<{ label: string; value: string | number; sub?: string; color: string }> = ({
  label, value, sub, color,
}) => (
  <div className={`p-4 rounded-lg border ${color}`}>
    <span className="block text-sm font-medium opacity-80">{label}</span>
    <span className="block text-2xl font-bold mt-1">{value}</span>
    {sub && <span className="block text-xs opacity-60 mt-1">{sub}</span>}
  </div>
);

const ProgressBar: React.FC<{ value: number; max: number; color: string }> = ({ value, max, color }) => {
  const pct = max > 0 ? Math.min(100, (value / max) * 100) : 0;
  return (
    <div className="w-full bg-gray-200 rounded-full h-2">
      <div className={`h-2 rounded-full ${color}`} style={{ width: `${pct}%` }} />
    </div>
  );
};

const ReportsPage: React.FC = () => {
  const { token } = useAuth();
  const [activeTab, setActiveTab] = useState<Tab>('summary');

  const [summary, setSummary] = useState<ReportSummary | null>(null);
  const [expiring, setExpiring] = useState<ExpirationAlert[]>([]);
  const [anonymized, setAnonymized] = useState<AnonymizedRecord[]>([]);
  const [performance, setPerformance] = useState<PerformanceMetrics | null>(null);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [daysAhead, setDaysAhead] = useState(30);

  const today = new Date();
  const yearStart = `${today.getFullYear()}-01-01`;
  const [anonStart, setAnonStart] = useState(yearStart);
  const [anonEnd, setAnonEnd] = useState(today.toISOString().slice(0, 10));

  const authHeaders = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' };

  const fetchSummary = useCallback(async () => {
    setLoading(true); setError(null);
    try {
      const r = await apiFetch('/reports/summary', { method: 'GET', headers: authHeaders });
      if (!r.ok) throw new Error('Failed to fetch summary');
      setSummary(await r.json());
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, [token]);

  const fetchExpiring = useCallback(async () => {
    setLoading(true); setError(null);
    try {
      const r = await apiFetch(`/reports/expiring?days=${daysAhead}`, { method: 'GET', headers: authHeaders });
      if (!r.ok) throw new Error('Failed to fetch expiration alerts');
      const data = await r.json();
      setExpiring(Array.isArray(data) ? data : []);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, [token, daysAhead]);

  const fetchAnonymized = useCallback(async () => {
    setLoading(true); setError(null);
    try {
      const start = Math.floor(new Date(anonStart).getTime() / 1000);
      const end = Math.floor(new Date(anonEnd).getTime() / 1000) + 86399;
      const r = await apiFetch(`/reports/anonymized?start=${start}&end=${end}`, { method: 'GET', headers: authHeaders });
      if (!r.ok) throw new Error('Failed to fetch anonymized records');
      const data = await r.json();
      setAnonymized(Array.isArray(data) ? data : []);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, [token, anonStart, anonEnd]);

  const fetchPerformance = useCallback(async () => {
    setLoading(true); setError(null);
    try {
      const r = await apiFetch('/reports/performance', { method: 'GET', headers: authHeaders });
      if (!r.ok) throw new Error('Failed to fetch performance metrics');
      setPerformance(await r.json());
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (activeTab === 'summary')     fetchSummary();
    if (activeTab === 'expiring')    fetchExpiring();
    if (activeTab === 'anonymized')  fetchAnonymized();
    if (activeTab === 'performance') fetchPerformance();
  }, [activeTab]);

  const tabs: { id: Tab; label: string }[] = [
    { id: 'summary',     label: 'Summary' },
    { id: 'expiring',    label: 'Expiration Alerts' },
    { id: 'anonymized',  label: 'Anonymized Data' },
    { id: 'performance', label: 'Performance' },
  ];

  return (
    <div className="container mx-auto pb-10">
      <h1 className="text-2xl font-semibold text-sky-700 mb-6">Reports</h1>

      {/* Tab bar */}
      <div className="flex gap-1 bg-gray-100 p-1 rounded-lg mb-6 w-fit">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === tab.id
                ? 'bg-white shadow-sm text-sky-700'
                : 'text-gray-500 hover:text-gray-700'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Loading / Error */}
      {loading && (
        <div className="flex justify-center items-center h-40">
          <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-sky-500" />
        </div>
      )}
      {!loading && error && (
        <div className="bg-red-50 border border-red-200 text-red-700 p-4 rounded-md">{error}</div>
      )}

      {/* --- SUMMARY TAB --- */}
      {!loading && !error && activeTab === 'summary' && summary && (
        <div className="space-y-6">
          {/* Key metrics */}
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
            <MetricCard label="Total Checks" value={summary.total_checks} color="bg-blue-50 text-blue-800 border-blue-200" />
            <MetricCard label="Last Month" value={summary.checks_last_month} color="bg-indigo-50 text-indigo-800 border-indigo-200" />
            <MetricCard label="This Month" value={summary.checks_this_month} color="bg-sky-50 text-sky-800 border-sky-200" />
            <MetricCard label="Expiring Next Month" value={summary.expiring_next_month.length} color="bg-orange-50 text-orange-800 border-orange-200" />
            <MetricCard label="Needs Review" value={summary.needs_review_count} sub="OCR confidence < 95%" color="bg-yellow-50 text-yellow-800 border-yellow-200" />
          </div>

          {/* Aviz breakdown */}
          <div className="bg-white rounded-lg shadow-sm p-6 text-gray-900">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-medium text-gray-800">Medical Opinion Breakdown</h2>
              <span className="text-xs text-gray-400">Only records where OCR parsed the aviz field</span>
            </div>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <MetricCard label="Fit" value={summary.fit_count} color="bg-green-50 text-green-800 border-green-200" />
              <MetricCard label="Conditionally Fit" value={summary.conditionally_fit_count} color="bg-teal-50 text-teal-800 border-teal-200" />
              <MetricCard label="Temporarily Unfit" value={summary.temporarily_unfit_count} color="bg-orange-50 text-orange-800 border-orange-200" />
              <MetricCard label="Unfit" value={summary.unfit_count} color="bg-red-50 text-red-800 border-red-200" />
            </div>
          </div>

          {/* OCR rate */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="bg-white rounded-lg shadow-sm p-6 text-gray-900">
              <h2 className="text-lg font-medium text-gray-800 mb-2">OCR Auto-Processing Rate</h2>
              <p className="text-3xl font-bold text-sky-700">{summary.ocr_success_rate.toFixed(1)}%</p>
              <p className="text-sm text-gray-500 mt-1">Records processed without manual review needed</p>
            </div>
            <div className="bg-white rounded-lg shadow-sm p-6 text-gray-900">
              <h2 className="text-lg font-medium text-gray-800 mb-2">Avg OCR Latency</h2>
              {summary.avg_processing_time_ms > 0 ? (
                <>
                  <p className="text-3xl font-bold text-indigo-700">{summary.avg_processing_time_ms} ms</p>
                  <p className="text-sm text-gray-500 mt-1">Average OCR + medical parsing time per image</p>
                </>
              ) : (
                <>
                  <p className="text-3xl font-bold text-gray-300">—</p>
                  <p className="text-sm text-gray-400 mt-1">No latency data yet — upload an image to populate</p>
                </>
              )}
            </div>
          </div>

          {/* Expiring next month preview */}
          {summary.expiring_next_month.length > 0 && (
            <div className="bg-white rounded-lg shadow-sm p-6 text-gray-900">
              <h2 className="text-lg font-medium text-gray-800 mb-4">Expiring Next Month ({summary.expiring_next_month.length})</h2>
              <table className="w-full text-sm text-left">
                <thead className="text-gray-500 border-b">
                  <tr>
                    <th className="pb-2 pr-4">Name</th>
                    <th className="pb-2 pr-4">Company</th>
                    <th className="pb-2 pr-4">Expires</th>
                    <th className="pb-2">Days Left</th>
                  </tr>
                </thead>
                <tbody>
                  {summary.expiring_next_month.map((a, i) => (
                    <tr key={i} className="border-b last:border-0">
                      <td className="py-2 pr-4 font-medium">{a.first_name} {a.last_name}</td>
                      <td className="py-2 pr-4 text-gray-600">{a.company || '—'}</td>
                      <td className="py-2 pr-4 text-gray-600">{fmtDate(a.next_exam_date)}</td>
                      <td className="py-2">
                        <span className={`px-2 py-0.5 rounded text-xs font-medium ${urgencyBadge(a.days_until_expiry)}`}>
                          {a.days_until_expiry}d
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* --- EXPIRATION ALERTS TAB --- */}
      {!loading && !error && activeTab === 'expiring' && (
        <div className="space-y-4">
          {/* Filter */}
          <div className="bg-white p-4 rounded-lg shadow-sm flex items-end gap-4 text-gray-900">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Days ahead</label>
              <select
                value={daysAhead}
                onChange={e => setDaysAhead(Number(e.target.value))}
                className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-sky-500 focus:outline-none"
              >
                <option value={7}>Next 7 days</option>
                <option value={30}>Next 30 days</option>
                <option value={60}>Next 60 days</option>
                <option value={90}>Next 90 days</option>
              </select>
            </div>
            <button
              onClick={fetchExpiring}
              className="px-4 py-2 bg-sky-600 text-white rounded-md text-sm hover:bg-sky-700"
            >
              Refresh
            </button>
          </div>

          {/* Legend */}
          <div className="flex gap-3 text-xs">
            <span className="px-2 py-1 rounded bg-red-100 text-red-700">≤ 7 days — Critical</span>
            <span className="px-2 py-1 rounded bg-orange-100 text-orange-700">≤ 30 days — Soon</span>
            <span className="px-2 py-1 rounded bg-yellow-100 text-yellow-700">&gt; 30 days — Upcoming</span>
          </div>

          {/* Table */}
          {expiring.length === 0 ? (
            <div className="bg-green-50 border border-green-200 text-green-700 p-6 rounded-lg text-center">
              No medical certificates expiring in the next {daysAhead} days.
            </div>
          ) : (
            <div className="bg-white rounded-lg shadow-sm overflow-hidden text-gray-900">
              <table className="w-full text-sm">
                <thead className="bg-gray-50 text-gray-500 border-b">
                  <tr>
                    <th className="text-left px-4 py-3">Employee</th>
                    <th className="text-left px-4 py-3">Company / Unit</th>
                    <th className="text-left px-4 py-3">Workplace</th>
                    <th className="text-left px-4 py-3">Expires</th>
                    <th className="text-left px-4 py-3">Days Left</th>
                    <th className="text-left px-4 py-3">Aviz</th>
                  </tr>
                </thead>
                <tbody>
                  {expiring
                    .sort((a, b) => a.days_until_expiry - b.days_until_expiry)
                    .map((alert, i) => (
                      <tr key={i} className={`border-b last:border-0 ${urgencyClass(alert.days_until_expiry)}`}>
                        <td className="px-4 py-3 font-medium">{alert.first_name} {alert.last_name}</td>
                        <td className="px-4 py-3">{alert.company || '—'}</td>
                        <td className="px-4 py-3">{alert.workplace || '—'}</td>
                        <td className="px-4 py-3">{fmtDate(alert.next_exam_date)}</td>
                        <td className="px-4 py-3">
                          <span className={`px-2 py-0.5 rounded text-xs font-bold ${urgencyBadge(alert.days_until_expiry)}`}>
                            {alert.days_until_expiry} days
                          </span>
                        </td>
                        <td className="px-4 py-3">{alert.medical_opinion || '—'}</td>
                      </tr>
                    ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* --- ANONYMIZED DATA TAB --- */}
      {!loading && !error && activeTab === 'anonymized' && (
        <div className="space-y-4">
          <div className="bg-amber-50 border border-amber-200 text-amber-800 p-3 rounded-md text-sm">
            PHI fields (Name, CNP, Personal ID) are excluded from this dataset for research/compliance use.
          </div>

          {/* Tabelul a fost mutat corect AICI, in interiorul tab-ului de date anonimizate */}
          {anonymized.length > 0 && (
            <div className="bg-white rounded-lg shadow-sm p-6 mb-2 text-gray-900">
              <h2 className="text-lg font-medium text-gray-800 mb-4">Fitness by Profession</h2>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {Object.entries(
                  anonymized.reduce((acc, curr) => {
                    const job = curr.job_title || 'Unknown';
                    if (!acc[job]) acc[job] = { fit: 0, unfit: 0, total: 0 };
                    acc[job].total += 1;
                    if (curr.medical_opinion === 'APT') acc[job].fit += 1;
                    else acc[job].unfit += 1;
                    return acc;
                  }, {} as Record<string, { fit: number; unfit: number; total: number }>)
                )
                  .sort(([, a], [, b]) => b.total - a.total)
                  .map(([job, stats], idx) => (
                    <div key={idx} className="border border-gray-100 rounded-md p-3 bg-gray-50">
                      <p className="font-semibold text-gray-700 truncate" title={job}>{job}</p>
                      <div className="flex justify-between mt-2 text-sm">
                        <span className="text-green-600 font-medium">{stats.fit} Fit</span>
                        <span className="text-red-500">{stats.unfit} Other</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-1.5 mt-2">
                        <div 
                          className="bg-green-500 h-1.5 rounded-full" 
                          style={{ width: `${(stats.fit / stats.total) * 100}%` }}
                        ></div>
                      </div>
                    </div>
                  ))}
              </div>
            </div>
          )}

          {/* Filters */}
          <div className="bg-white p-4 rounded-lg shadow-sm flex flex-wrap items-end gap-4 text-gray-900">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
              <input
                type="date"
                value={anonStart}
                onChange={e => setAnonStart(e.target.value)}
                className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-sky-500 focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">End Date</label>
              <input
                type="date"
                value={anonEnd}
                onChange={e => setAnonEnd(e.target.value)}
                className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-sky-500 focus:outline-none"
              />
            </div>
            <button
              onClick={fetchAnonymized}
              className="px-4 py-2 bg-sky-600 text-white rounded-md text-sm hover:bg-sky-700"
            >
              Apply
            </button>
          </div>

          {anonymized.length === 0 ? (
            <div className="text-center text-gray-500 py-10">No records found for this period.</div>
          ) : (
            <div className="bg-white rounded-lg shadow-sm overflow-x-auto text-gray-900">
              <table className="w-full text-sm">
                <thead className="bg-gray-50 text-gray-500 border-b">
                  <tr>
                    <th className="text-left px-4 py-3">Control Type</th>
                    <th className="text-left px-4 py-3">Medical Opinion</th>
                    <th className="text-left px-4 py-3">Company</th>
                    <th className="text-left px-4 py-3">Role</th>
                    <th className="text-left px-4 py-3">Date</th>
                    <th className="text-left px-4 py-3">Next Exam</th>
                    <th className="text-left px-4 py-3">OCR Conf.</th>
                  </tr>
                </thead>
                <tbody>
                  {anonymized.map((rec, i) => (
                    <tr key={i} className="border-b last:border-0 hover:bg-gray-50">
                      <td className="px-4 py-2">{rec.control_type || '—'}</td>
                      <td className="px-4 py-2">
                        <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                          rec.medical_opinion === 'APT' ? 'bg-green-100 text-green-700' :
                          rec.medical_opinion === 'APT CONDITIONAT' ? 'bg-teal-100 text-teal-700' :
                          rec.medical_opinion === 'INAPT TEMPORAR' ? 'bg-orange-100 text-orange-700' :
                          rec.medical_opinion === 'INAPT' ? 'bg-red-100 text-red-700' :
                          'bg-gray-100 text-gray-600'
                        }`}>
                          {rec.medical_opinion || '—'}
                        </span>
                      </td>
                      <td className="px-4 py-2 text-gray-600">{rec.company || '—'}</td>
                      <td className="px-4 py-2 text-gray-600">{rec.job_title || '—'}</td>
                      <td className="px-4 py-2 text-gray-600">{fmtDate(rec.exam_date)}</td>
                      <td className="px-4 py-2 text-gray-600">{fmtDate(rec.next_exam_date)}</td>
                      <td className="px-4 py-2">
                        <span className={`text-xs font-medium ${
                          rec.ocr_confidence >= 95 ? 'text-green-600' : 'text-orange-600'
                        }`}>
                          {rec.ocr_confidence.toFixed(1)}%
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <div className="px-4 py-2 text-xs text-gray-400 border-t">{anonymized.length} records</div>
            </div>
          )}
        </div>
      )}

      {/* --- PERFORMANCE TAB --- */}
      {!loading && !error && activeTab === 'performance' && performance && (
        <div className="space-y-6">
          <div className="bg-white rounded-lg shadow-sm p-6 space-y-6 text-gray-900">
            <h2 className="text-lg font-medium text-gray-800">System Performance Metrics</h2>

            <div>
              <div className="flex justify-between text-sm mb-1">
                <span className="text-gray-600">Auto-processing success rate</span>
                <span className="font-semibold text-sky-700">{performance.ocr_success_rate.toFixed(1)}%</span>
              </div>
              <ProgressBar value={performance.ocr_success_rate} max={100} color="bg-sky-500" />
            </div>

            <div>
              <div className="flex justify-between text-sm mb-1">
                <span className="text-gray-600">Average OCR confidence</span>
                <span className={`font-semibold ${performance.avg_ocr_confidence >= 95 ? 'text-green-600' : 'text-orange-600'}`}>
                  {performance.avg_ocr_confidence.toFixed(1)}%
                </span>
              </div>
              <ProgressBar value={performance.avg_ocr_confidence} max={100} color={performance.avg_ocr_confidence >= 95 ? 'bg-green-500' : 'bg-orange-400'} />
            </div>

            <div>
              <div className="flex justify-between text-sm mb-1">
                <span className="text-gray-600">Manual reviews completed</span>
                <span className="font-semibold text-gray-700">
                  {performance.reviewed_count} / {performance.needs_review_count}
                </span>
              </div>
              <ProgressBar
                value={performance.reviewed_count}
                max={performance.needs_review_count}
                color="bg-teal-500"
              />
            </div>
          </div>

          {/* Processing latency */}
          {performance.avg_processing_time_ms > 0 && (
            <div className="bg-white rounded-lg shadow-sm p-6 text-gray-900">
              <h2 className="text-lg font-medium text-gray-800 mb-4">Processing Latency (OCR + Parsing)</h2>
              <div className="grid grid-cols-3 gap-4">
                <div className="bg-sky-50 border border-sky-100 rounded-lg p-4 text-center">
                  <span className="block text-2xl font-bold text-sky-700">{performance.avg_processing_time_ms} ms</span>
                  <span className="block text-sm text-sky-600 mt-1">Average</span>
                </div>
                <div className="bg-green-50 border border-green-100 rounded-lg p-4 text-center">
                  <span className="block text-2xl font-bold text-green-700">{performance.min_processing_time_ms} ms</span>
                  <span className="block text-sm text-green-600 mt-1">Minimum</span>
                </div>
                <div className="bg-orange-50 border border-orange-100 rounded-lg p-4 text-center">
                  <span className="block text-2xl font-bold text-orange-700">{performance.max_processing_time_ms} ms</span>
                  <span className="block text-sm text-orange-600 mt-1">Maximum</span>
                </div>
              </div>
              <p className="text-xs text-gray-400 mt-3">Measured from image receipt to OCR + medical field extraction completion. Only photos processed after latency tracking was enabled are included.</p>
            </div>
          )}

        </div>
      )}
    </div>
  );
};

export default ReportsPage;