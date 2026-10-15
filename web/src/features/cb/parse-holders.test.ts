import { describe, expect, it } from 'vitest'
import { parseHoldersPaste } from './parse-holders'

describe('parseHoldersPaste', () => {
  it('parses tab-separated broker paste', () => {
    const raw = `截止日期2026-06-30
持有人名称	持有数量(张)	持有比例(%)	环比
杨九红	255.7万	44.08	不变
陈定红	109.6万	18.89	不变
可转债ETF博时	53.93万	9.30	新进
华宝可转债债券A	23.00万	3.97	新进
常州途朗投资合伙企业(有限合伙)	20.29万	3.50	不变
常州途阳投资合伙企业(有限合伙)	20.29万	3.50	不变
方正证券股份有限公司	8.739万	1.51	新进
中信证券股份有限公司	6.600万	1.14	新进
中银国有企业债A	4.162万	0.72	新进
中邮证券有限责任公司	4.000万	0.69	新进
合计	506.2万	87.29	--
`
    const p = parseHoldersPaste(raw)
    expect(p.end_date).toBe('2026-06-30')
    expect(p.holders).toHaveLength(10)
    expect(p.holders[0]).toEqual({
      holder_rank: 1,
      holder_name: '杨九红',
      hold_amount: 255.7,
      hold_ratio: 44.08,
    })
    expect(p.holders[9]).toMatchObject({
      holder_rank: 10,
      holder_name: '中邮证券有限责任公司',
      hold_amount: 4,
      hold_ratio: 0.69,
    })
  })

  it('rejects missing cutoff date', () => {
    expect(() => parseHoldersPaste('持有人名称\t数量\t比例\n甲\t1万\t1')).toThrow(/截止日期/)
  })
})
