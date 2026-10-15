/** 粘贴解析结果。hold_amount 单位：万张。 */
export type ParsedHolder = {
  holder_rank: number
  holder_name: string
  hold_amount: number
  hold_ratio: number
}

export type ParsedHolders = {
  end_date: string // YYYY-MM-DD
  holders: ParsedHolder[]
}

const reCutoff = /截止日期\s*[:：]?\s*(\d{4}[-/]?\d{2}[-/]?\d{2})/
const reAmount = /^\s*([0-9]+(?:\.[0-9]+)?)\s*万?\s*$/

function splitCols(line: string): string[] {
  if (line.includes('\t')) {
    const parts = line.split('\t')
    const out: string[] = []
    for (const p of parts) {
      const t = p.trim()
      if (t !== '' || out.length > 0) out.push(t)
    }
    while (out.length > 0 && out[out.length - 1] === '') out.pop()
    return out
  }
  return line.trim().split(/\s+/).filter(Boolean)
}

function parseLooseDate(s: string): string {
  const raw = s.trim().replace(/\//g, '-')
  if (raw.includes('-')) {
    const m = /^(\d{4})-(\d{1,2})-(\d{1,2})$/.exec(raw)
    if (!m) throw new Error(`无法解析日期 "${s}"`)
    const y = m[1]
    const mo = m[2].padStart(2, '0')
    const d = m[3].padStart(2, '0')
    return `${y}-${mo}-${d}`
  }
  if (raw.length === 8 && /^\d{8}$/.test(raw)) {
    return `${raw.slice(0, 4)}-${raw.slice(4, 6)}-${raw.slice(6, 8)}`
  }
  throw new Error(`无法解析日期 "${s}"`)
}

function parseAmountWan(s: string): number {
  const cleaned = s.trim().replace(/,/g, '')
  const m = reAmount.exec(cleaned)
  if (!m) throw new Error(`期望如 33.77万，得到 ${JSON.stringify(s)}`)
  return Number(m[1])
}

function parseRatio(s: string): number {
  const cleaned = s.trim().replace(/%$/, '').replace(/,/g, '')
  const n = Number(cleaned)
  if (Number.isNaN(n)) throw new Error(`比例无效 ${JSON.stringify(s)}`)
  return n
}

/** 解析券商/网页复制的十大持有人文本。 */
export function parseHoldersPaste(text: string): ParsedHolders {
  const normalized = text.replace(/\r\n/g, '\n').replace(/\r/g, '\n')
  const lines = normalized.split('\n')

  let endDate = ''
  const holders: ParsedHolder[] = []
  let rank = 0

  for (const raw of lines) {
    const line = raw.trim()
    if (!line) continue

    const cutoff = reCutoff.exec(line)
    if (cutoff) {
      endDate = parseLooseDate(cutoff[1])
      continue
    }
    if (line.includes('持有人名称')) continue
    if (line.startsWith('合计')) continue

    const cols = splitCols(line)
    if (cols.length < 3) {
      throw new Error(`无法解析行（至少需要名称、数量、比例）: ${line}`)
    }
    const name = cols[0].trim()
    if (!name) throw new Error(`持有人名称为空: ${line}`)

    let amount: number
    let ratio: number
    try {
      amount = parseAmountWan(cols[1])
    } catch (e) {
      throw new Error(`数量无效 ${JSON.stringify(cols[1])}: ${e instanceof Error ? e.message : e}`)
    }
    try {
      ratio = parseRatio(cols[2])
    } catch (e) {
      throw new Error(`比例无效 ${JSON.stringify(cols[2])}: ${e instanceof Error ? e.message : e}`)
    }

    rank += 1
    holders.push({
      holder_rank: rank,
      holder_name: name,
      hold_amount: amount,
      hold_ratio: ratio,
    })
  }

  if (!endDate) {
    throw new Error('未找到截止日期，请包含「截止日期2026-06-30」一行')
  }
  if (holders.length === 0) {
    throw new Error('未解析到任何持有人行')
  }
  return { end_date: endDate, holders }
}
