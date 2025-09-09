import re
import requests
import json
import pandas as pd
import hashlib
import urllib.parse
import time
import sqlite3
import os

# 数据库连接
def get_db_connection():
    db_path = './data/crawler.db'
    if not os.path.exists(os.path.dirname(db_path)):
        os.makedirs(os.path.dirname(db_path))
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    # 如果表不存在，则创建 bilibili_comments 表
    cursor.execute("""
    CREATE TABLE IF NOT EXISTS bilibili_comments (
        序号 INTEGER,
        上级评论ID INTEGER,
        评论ID INTEGER PRIMARY KEY,
        用户ID INTEGER,
        用户名 TEXT,
        用户等级 INTEGER,
        性别 TEXT,
        评论内容 TEXT,
        评论时间 TEXT,
        回复数 INTEGER,
        点赞数 INTEGER,
        个性签名 TEXT,
        IP属地 TEXT,
        是否是大会员 TEXT,
        头像 TEXT,
        视频BV号 TEXT,
        视频标题 TEXT
    )
    """)
    conn.commit()
    return conn

# 插入评论到数据库
def insert_comment_to_db(conn, comment_data):
    cursor = conn.cursor()
    sql = '''
    INSERT OR IGNORE INTO bilibili_comments 
    (序号, 上级评论ID, 评论ID, 用户ID, 用户名, 用户等级, 性别, 评论内容, 评论时间, 回复数, 点赞数, 个性签名, IP属地, 是否是大会员, 头像, 视频BV号, 视频标题)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    '''
    cursor.execute(sql, comment_data)
    conn.commit()

# 获取B站的Header
def get_Header():
    with open('bili_cookie.txt','r') as f:
            cookie=f.read()
    header={
            "Cookie":cookie,
            "User-Agent":'Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:135.0) Gecko/20100101 Firefox/135.0'
    }
    return header

# 通过bv号，获取视频的oid
def get_information(bv):
    resp = requests.get(f"https://www.bilibili.com/video/{bv}/?p=14&spm_id_from=pageDriver&vd_source=cd6ee6b033cd2da64359bad72619ca8a",headers=get_Header())
    # 提取视频oid
    obj = re.compile(f'"aid":(?P<id>.*?),"bvid":"{bv}"')
    oid = obj.search(resp.text).group('id') # type: ignore

    # 提取视频的标题
    obj = re.compile(r'<title>(?P<title>.*?)</title>')
    try:
        title = obj.search(resp.text).group('title') # type: ignore
    except:
        title = "未识别"

    return oid,title

# MD5加密
def md5(code):
    MD5 = hashlib.md5()
    MD5.update(code.encode('utf-8'))
    w_rid = MD5.hexdigest()
    return w_rid

# 轮页爬取
def start(bv, oid, pageID, count, conn, title, is_second):
    # 参数
    mode = 2   # 为2时爬取的是最新评论，为3时爬取的是热门评论
    plat = 1
    type = 1  
    web_location = 1315875

    # 获取当下时间戳
    wts = int(time.time())
    
    # 如果不是第一页
    if pageID != '':
        pagination_str = '{"offset":"%s"}' % pageID
        code = f"mode={mode}&oid={oid}&pagination_str={urllib.parse.quote(pagination_str)}&plat={plat}&type={type}&web_location={web_location}&wts={wts}" + 'ea1db124af3c7062474693fa704f4ff8'
        w_rid = md5(code)
        url = f"https://api.bilibili.com/x/v2/reply/wbi/main?oid={oid}&type={type}&mode={mode}&pagination_str={urllib.parse.quote(pagination_str, safe=':')}&plat=1&web_location=1315875&w_rid={w_rid}&wts={wts}"
    
    # 如果是第一页
    else:
        pagination_str = '{"offset":""}'
        code = f"mode={mode}&oid={oid}&pagination_str={urllib.parse.quote(pagination_str)}&plat={plat}&type={type}&web_location={web_location}&wts={wts}" + 'ea1db124af3c7062474693fa704f4ff8'
        w_rid = md5(code)
        url = f"https://api.bilibili.com/x/v2/reply/wbi/main?oid={oid}&type={type}&mode={mode}&pagination_str={urllib.parse.quote(pagination_str, safe=':')}&plat=1&web_location=1315875&w_rid={w_rid}&wts={wts}"
    

    comment = requests.get(url=url, headers=get_Header()).content.decode('utf-8')
    comment = json.loads(comment)

    for reply in comment['data']['replies']:
        # 评论数量+1
        count += 1

        if count % 1000 ==0:
            time.sleep(20)

        # 上级评论ID
        parent=reply["parent"]
        # 评论ID
        rpid = reply["rpid"]
        # 用户ID
        uid = reply["mid"]
        # 用户名
        name = reply["member"]["uname"]
        # 用户等级
        level = reply["member"]["level_info"]["current_level"]
        # 性别
        sex = reply["member"]["sex"]
        # 头像
        avatar = reply["member"]["avatar"]
        # 是否是大会员
        if reply["member"]["vip"]["vipStatus"] == 0:
            vip = "否"
        else:
            vip = "是"
        # IP属地
        try:
            IP = reply["reply_control"]['location'][5:]
        except:
            IP = "未知"
        # 内容
        context = reply["content"]["message"]
        # 评论时间
        reply_time = pd.to_datetime(reply["ctime"], unit='s').strftime('%Y-%m-%d %H:%M:%S')
        # 相关回复数
        try:
            rereply = reply["reply_control"]["sub_reply_entry_text"]
            rereply = int(re.findall(r'\d+', rereply)[0])
        except:
            rereply = 0
        # 点赞数
        like = reply['like']

        # 个性签名
        try:
            sign = reply['member']['sign']
        except:
            sign = ''

        # 写入数据库
        comment_data = (count, parent, rpid, uid, name, level, sex, context, reply_time, rereply, like, sign, IP, vip, avatar, bv, title)
        insert_comment_to_db(conn, comment_data)

        # 二级评论(如果开启了二级评论爬取，且该评论回复数不为0，则爬取该评论的二级评论)
        if is_second and rereply !=0:
            for page in range(1,rereply//10+2):
                second_url=f"https://api.bilibili.com/x/v2/reply/reply?oid={oid}&type=1&root={rpid}&ps=10&pn={page}&web_location=333.788"
                second_comment=requests.get(url=second_url,headers=get_Header()).content.decode('utf-8')
                second_comment=json.loads(second_comment)
                for second in second_comment['data']['replies']:
                    # 评论数量+1
                    count += 1
                    # 上级评论ID
                    parent=second["parent"]
                    # 评论ID
                    second_rpid = second["rpid"]
                    # 用户ID
                    uid = second["mid"]
                    # 用户名
                    name = second["member"]["uname"]
                    # 用户等级
                    level = second["member"]["level_info"]["current_level"]
                    # 性别
                    sex = second["member"]["sex"]
                    # 头像
                    avatar = second["member"]["avatar"]
                    # 是否是大会员
                    if second["member"]["vip"]["vipStatus"] == 0:
                        vip = "否"
                    else:
                        vip = "是"
                    # IP属地
                    try:
                        IP = second["reply_control"]['location'][5:]
                    except:
                        IP = "未知"
                    # 内容
                    context = second["content"]["message"]
                    # 评论时间
                    reply_time = pd.to_datetime(second["ctime"], unit='s').strftime('%Y-%m-%d %H:%M:%S')
                    # 相关回复数
                    try:
                        rereply = second["reply_control"]["sub_reply_entry_text"]
                        rereply = re.findall(r'\d+', rereply)[0]
                    except:
                        rereply = 0
                    # 点赞数
                    like = second['like']
                    # 个性签名
                    try:
                        sign = second['member']['sign']
                    except:
                        sign = ''

                    # 写入数据库
                    comment_data = (count, parent, second_rpid, uid, name, level, sex, context, reply_time, rereply, like, sign, IP, vip, avatar, bv, title)
                    insert_comment_to_db(conn, comment_data)
            


    # 下一页的pageID
    try:
        next_pageID = comment['data']['cursor']['pagination_reply']['next_offset']
    except:
        next_pageID = 0

    # 判断是否是最后一页了
    if next_pageID == 0:
        print(f"评论爬取完成！总共爬取{count}条。")
        return bv, oid, next_pageID, count, conn, is_second
    # 如果不是最后一页，则停0.5s（避免反爬机制）
    else:
        time.sleep(0.5)
        print(f"当前爬取{count}条。")
        return bv, oid, next_pageID, count, conn, is_second

if __name__ == "__main__":


    # 获取视频bv
    bv = "BV1HW4y1n7BF"
    # 获取视频oid和标题
    oid,title = get_information(bv)
    # 评论起始页（默认为空）
    next_pageID = ''
    # 初始化评论数量
    count = 0


    # 是否开启二级评论爬取，默认开启
    is_second = True


    # 连接数据库
    conn = get_db_connection()
    print(f"开始爬取视频 {bv} 的评论，标题：{title}")
    
    try:
        # 开始爬取
        while next_pageID != 0:
            bv, oid, next_pageID, count, conn, is_second = start(bv, oid, next_pageID, count, conn, title, is_second)
    finally:
        # 确保数据库连接关闭
        conn.close()
        print(f"数据库连接已关闭，所有评论已保存到 SQLite 数据库中。")
