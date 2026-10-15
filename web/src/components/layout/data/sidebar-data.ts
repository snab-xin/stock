import {
  CandlestickChart,
  Database,
  Landmark,
  LineChart,
} from 'lucide-react'
import { type SidebarData } from '../types'

export const sidebarData: SidebarData = {
  user: {
    name: 'local',
    email: 'ashare@local',
    avatar: '/avatars/shadcn.jpg',
  },
  teams: [
    {
      name: 'ashare',
      logo: CandlestickChart,
      plan: 'A 股数据看板',
    },
  ],
  navGroups: [
    {
      title: '业务',
      items: [
        {
          title: '行情质检',
          url: '/',
          icon: CandlestickChart,
        },
        {
          title: '可转债',
          url: '/cb',
          icon: Landmark,
        },
        {
          title: '采集任务',
          url: '/ingest',
          icon: Database,
        },
        {
          title: '最小回测',
          url: '/backtest',
          icon: LineChart,
        },
      ],
    },
  ],
}
