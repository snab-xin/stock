package tushare

// 请求 Tushare 时显式列出全部输出字段，避免默认字段漏列。

const stockBasicFields = "ts_code,symbol,name,area,industry,fullname,enname,cnspell,market," +
	"exchange,curr_type,list_status,list_date,delist_date,is_hs,act_name,act_ent_type"

const tradeCalFields = "exchange,cal_date,is_open,pretrade_date"

const dailyFields = "ts_code,trade_date,open,high,low,close,pre_close,change,pct_chg,vol,amount,ah_vol,ah_amount"

const stkLimitFields = "trade_date,ts_code,pre_close,up_limit,down_limit,asset_type,exchange"

const dailyBasicFields = "ts_code,trade_date,close,turnover_rate,turnover_rate_f,volume_ratio," +
	"pe,pe_ttm,pb,ps,ps_ttm,dv_ratio,dv_ttm,total_share,float_share,free_share,total_mv,circ_mv,limit_status"

const adjFactorFields = "ts_code,trade_date,adj_factor"

const suspendFields = "ts_code,trade_date,suspend_timing,suspend_type"

const indexBasicFields = "ts_code,name,fullname,market,publisher,index_type,category," +
	"base_date,base_point,list_date,weight_rule,desc,exp_date"

const indexDailyFields = "ts_code,trade_date,close,open,high,low,pre_close,change,pct_chg,vol,amount"

const indexClassifyFields = "index_code,industry_name,parent_code,level,industry_code,is_pub,src"

const swDailyFields = "ts_code,trade_date,name,open,low,high,close,change,pct_change,vol,amount,pe,pb,float_mv,total_mv"

const etfBasicFields = "ts_code,csname,extname,cname,index_code,index_name,setup_date,list_date," +
	"list_status,exchange,mgr_name,custod_name,mgt_fee,etf_type"

const fundDailyFields = "ts_code,trade_date,open,high,low,close,pre_close,change,pct_chg,vol,amount"

const fundAdjFields = "ts_code,trade_date,adj_factor"

const etfShareSizeFields = "trade_date,ts_code,etf_name,total_share,total_size,nav,close,exchange"

const etfSHConsFields = "trade_date,ts_code,con_code,con_name,qty,sub_flag,cpr,rdr,sca,exchange"

const etfSZConsFields = "trade_date,ts_code,con_code,con_name,qty,sub_flag,cpr,rdr,sub_cc,red_cc,exchange"
