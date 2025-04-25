#!/bin/bash


domains=("google.com" "facebook.com" "youtube.com" "baidu.com" "yahoo.com" "amazon.com" "wikipedia.org" "qq.com" "twitter.com" "slashdot.org" "google.co.in" "taobao.com" "live.com" "sina.com.cn" "yahoo.co.jp" "linkedin.com" "weibo.com" "ebay.com" "google.co.jp" "yandex.ru" "bing.com" "vk.com" "hao123.com" "google.de" "instagram.com" "t.co" "msn.com" "amazon.co.jp" "tmall.com" "google.co.uk" "pinterest.com" "ask.com" "reddit.com" "wordpress.com" "mail.ru" "google.fr" "blogspot.com" "paypal.com" "onclickads.net" "google.com.br" "tumblr.com" "apple.com" "google.ru" "aliexpress.com" "sohu.com" "microsoft.com" "imgur.com" "xvideos.com" "google.it" "imdb.com" "google.es" "netflix.com" "gmw.cn" "amazon.de" "fc2.com" "360.cn" "alibaba.com" "go.com" "stackoverflow.com" "ok.ru" "google.com.mx" "google.ca" "amazon.in" "google.com.hk" "tianya.cn" "amazon.co.uk" "craigslist.org" "rakuten.co.jp" "naver.com" "blogger.com" "diply.com" "google.com.tr" "xhamster.com" "flipkart.com" "espn.go.com" "soso.com" "outbrain.com" "nicovideo.jp" "google.co.id" "cnn.com" "xinhuanet.com" "dropbox.com" "google.co.kr" "googleusercontent.com" "github.com" "bongacams.com" "ebay.de" "kat.cr" "bbc.co.uk" "google.pl" "google.com.au" "pixnet.net" "tradeadexchange.com" "popads.net" "googleadservices.com" "ebay.co.uk" "dailymotion.com" "sogou.com" "adnetworkperformance.com" "adobe.com" "directrev.com" "nytimes.com" "jd.com" "wikia.com" "adcash.com" "livedoor.jp" "booking.com" "163.com" "bbc.com" "alipay.com" "coccoc.com" "dailymail.co.uk" "indiatimes.com" "china.com" "dmm.co.jp" "china.com.cn" "chase.com" "buzzfeed.com" "google.com.sa" "huffingtonpost.com" "youku.com" "google.com.eg" "google.com.tw" "terraclicks.com" "uol.com.br" "amazon.cn" "snapdeal.com" "office.com" "google.com.ar" "microsoftonline.com" "walmart.com" "ameblo.jp" "amazon.fr" "daum.net" "amazonaws.com" "blogspot.in" "slideshare.net" "etsy.com" "twitch.tv" "google.com.pk" "whatsapp.com" "bankofamerica.com" "yelp.com" "globo.com" "theguardian.com" "tudou.com" "flickr.com" "aol.com" "stackexchange.com" "chinadaily.com.cn" "cnet.com" "weather.com" "indeed.com" "ettoday.net" "amazon.it" "reimageplus.com" "quora.com" "redtube.com" "soundcloud.com" "detail.tmall.com" "google.nl" "forbes.com" "douban.com" "loading-delivery2.com" "naver.jp" "bp.blogspot.com" "cntv.cn" "cnzz.com" "google.co.za" "wellsfargo.com" "google.co.ve" "target.com" "adf.ly" "zillow.com" "vice.com" "google.gr" "leboncoin.fr" "kakaku.com" "ikea.com" "gmail.com" "bestbuy.com" "vimeo.com" "avito.ru" "godaddy.com" "spaceshipads.com" "goo.ne.jp" "salesforce.com" "about.com" "tripadvisor.com" "allegro.pl" "livejournal.com" "nih.gov" "tubecup.com" "adplxmd.com" "foxnews.com" "deviantart.com" "files.wordpress.com" "doublepimp.com" "google.com.ua" "washingtonpost.com" "theladbible.com" "w3schools.com" "themeforest.net" "feedly.com" "wikihow.com" "wordpress.org" "office365.com" "taboola.com" "9gag.com" "mozilla.org" "akamaihd.net" "zol.com.cn" "hclips.com" "mediafire.com" "businessinsider.com" "google.cn" "onet.pl" "comcast.net" "gfycat.com" "softonic.com" "google.com.co" "pixiv.net" "google.co.th" "zhihu.com" "americanexpress.com" "amazon.es" "mystart.com" "nfl.com" "wix.com" "steamcommunity.com" "archive.org" "usps.com" "ups.com" "google.com.sg" "wikimedia.org" "bilibili.com" "homedepot.com" "google.ro" "secureserver.net" "doorblog.jp" "force.com" "telegraph.co.uk" "skype.com" "detik.com" "shutterstock.com" "google.com.ng" "ebay-kleinanzeigen.de" "weebly.com" "popcash.net" "google.com.ph" "addthis.com" "steampowered.com" "web.de" "bitauto.com" "blogspot.com.br" "google.se" "github.io" "rambler.ru" "avg.com" "ndtv.com" "hulu.com" "gamer.com.tw" "xywy.com" "huanqiu.com" "nametests.com" "51.la" "orange.fr" "tlbb8.com" "sourceforge.net" "hdfcbank.com" "livejasmin.com" "espncricinfo.com" "answers.com" "hp.com" "gmx.net" "youm7.com" "mailchimp.com" "mercadolivre.com.br" "speedtest.net" "xfinity.com" "ebay.in" "webmd.com" "ifeng.com" "google.at" "groupon.com" "blogfa.com" "wordreference.com" "uptodown.com" "xuite.net" "media.tumblr.com" "hootsuite.com" "usatoday.com" "google.pt" "capitalone.com" "stumbleupon.com" "goodreads.com" "wp.pl" "people.com.cn" "bet365.com" "google.be" "t-online.de" "paytm.com" "fedex.com" "fbcdn.net" "icicibank.com" "blog.jp" "google.com.pe" "thesaurus.com" "bloomberg.com" "mashable.com" "caijing.com.cn" "bild.de" "extratorrent.cc" "warmportrait.com" "dmm.com" "pandora.com" "putlocker.is" "amazon.ca" "spiegel.de" "seznam.cz" "google.ae" "spotify.com" "wsj.com" "dell.com" "ign.com" "jabong.com" "udn.com" "2ch.net" "macys.com" "chaturbate.com" "kaskus.co.id" "att.com" "engadget.com" "accuweather.com" "gameforge.com" "varzesh3.com" "watsons.tmall.com" "life.com.tw" "smzdm.com" "badoo.com" "google.ch" "mama.cn" "samsung.com" "adidas.tmall.com" "rutracker.org" "1688.com" "chaoshi.tmall.com" "1905.com" "gsmarena.com" "google.az" "youth.cn" "onlinesbi.com" "styletv.com.cn" "abs-cbnnews.com" "mega.nz" "twimg.com" "liveadexchanger.com" "livedoor.biz" "zendesk.com" "trello.com" "mlb.com" "rediff.com" "tistory.com" "39.net" "reference.com" "google.cl" "google.com.bd" "google.cz" "milliyet.com.tr" "reuters.com" "icloud.com" "verizonwireless.com" "haosou.com" "liputan6.com" "kohls.com" "kickstarter.com" "kouclo.com" "sahibinden.com" "shopclues.com" "enet.com.cn" "ebay.it" "mydomainadvisor.com" "iqiyi.com" "sberbank.ru" "impress.co.jp" "eksisozluk.com" "bleacherreport.com" "slickdeals.net" "yaolan.com" "tube8.com" "evernote.com" "trackingclick.net" "babytree.com" "baike.com" "lady8844.com" "infusionsoft.com" "hurriyet.com.tr" "ask.fm" "google.hu" "liveinternet.ru" "flirchi.com" "newegg.com" "ijreview.com" "torrentz.eu" "vid.me" "likes.com" "kinopoisk.ru" "thefreedictionary.com" "youradexchange.com" "pinimg.com" "oracle.com" "ppomppu.co.kr" "google.ie" "gap.com" "4shared.com" "rt.com" "google.co.il" "yandex.ua" "scribd.com" "ebay.com.au" "quikr.com" "photobucket.com" "ltn.com.tw" "taleo.net" "repubblica.it" "ce.cn" "libero.it" "onedio.com" "list-manage.com" "uploaded.net" "slack.com" "blogspot.com.es" "blogimg.jp" "livedoor.com" "meetup.com" "cbssports.com" "retailmenot.com" "goal.com" "goodgamestudios.com" "cnnic.cn" "eastday.com" "citi.com" "lifehacker.com" "51yes.com" "exoclick.com" "buzzfil.net" "olx.in" "hm.com" "neobux.com" "ameba.jp" "cloudfront.net" "teepr.com" "pconline.com.cn" "google.dz" "kinogo.co" "gizmodo.com" "elpais.com" "savefrom.net" "rbc.ru" "disqus.com" "fiverr.com" "theverge.com" "ewt.cc" "marca.com" "xda-developers.com" "lowes.com" "free.fr" "google.fi" "allrecipes.com" "xe.com" "battle.net" "torrentz.in" "kompas.com" "surveymonkey.com" "aparat.com" "souq.com" "ilividnewtab.com" "mobile.de" "nordstrom.com" "stockstar.com" "nyaa.se" "time.com" "asos.com" "intuit.com" "youboy.com" "nbcnews.com" "naukri.com" "4dsply.com" "epweike.com" "streamcloud.eu" "techcrunch.com" "medium.com" "tabelog.com" "independent.co.uk" "chip.de" "zippyshare.com" "lenovo.com" "expedia.com" "wunderground.com" "java.com")
domans_length=${#domains[@]}
dns_server_ip=${GOAWAY_IP}
dns_server_port=${GOAWAY_PORT:-53}

if [ -z "${dns_server_ip}" ]; then
    echo "GOAWAY_IP not set, quitting."
    exit 1
fi

success_count=0
blocked_count=0
total_time=0
lockfile="/tmp/dns_benchmark_lock"

query_domain() {
    local domain=$1
    local start_time end_time duration result
    start_time=$(date +%s%3N)
    result=$(dig +short @$dns_server_ip -p $dns_server_port $domain)
    end_time=$(date +%s%3N)

    duration=$((end_time - start_time))
    echo "$duration" >> $lockfile

    if [[ "$result" == "0.0.0.0" || -z "$result" ]]; then
        echo "Query for $domain blocked in ${duration}ms."
        echo "blocked" >> $lockfile
    else
        echo "Query for $domain succeeded in ${duration}ms."
        echo "success" >> $lockfile
    fi
}

> $lockfile

echo "Sending ${#domains[@]} requests to $dns_server_ip on port $dns_server_port..."

for domain in "${domains[@]}"; do
    query_domain "$domain" &
done

wait

while read -r line; do
    if [[ "$line" == "success" ]]; then
        ((success_count++))
    elif [[ "$line" == "blocked" ]]; then
        ((blocked_count++))
    elif [[ "$line" =~ ^[0-9]+$ ]]; then
        total_time=$((total_time + line))
    fi
done < $lockfile

# Cleanup
rm -f $lockfile

average_time=$((total_time / ${#domains[@]}))

echo
echo "### Summary ###"
echo "Domains:      $domans_length"
echo "Success:      $success_count"
echo "Blocked:      $blocked_count"
echo "Total time:   ${total_time} ms."
echo "Average time: ${average_time} ms."
