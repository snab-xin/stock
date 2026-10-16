import { parseISO, format } from 'date-fns'

export function formatISODateTimeStr(date: string) {
    if(date === null || date === undefined) {
        return null
    }
    const d = parseISO(date)
    return format(d, 'yyyy-MM-dd HH:mm:ss')
}