#!/bin/bash

domains=("0-02.net" "000free.us" "000nethost.com" "000space.com" "000webhost.info" "000webhostapp.com" "001union.com" "002777.xyz" "003store.com" "0063450bb8e5e3369ef8c167a463d15c01527283b79bd266bb19b23e.com" "fc96c51f92.0074854f80.com" "00771944.xyz" "007angels.com" "007arcadegames.com" "analytics.007computer.com" "007guard.com" "007itshop.com" "cdn.007moms.com" "007sites.com" "0086market.com" "31862bf36c.008dd84707.com" "00author.com" "00d2e2e5ef.com" "00d3ed994e.com" "00ebjdbagyqwt.club" "b724ebdb0a.00f031b898.com" "00f8c4bb25.com" "00fun.com" "00go.com" "00hq.com" "00it.com" "00px.net" "00sexus.com" "00webcams.com" "00xvppy.cn" "analytics.0101.fr" "010172.com" "01045395.xyz" "bd100.010lm.com" "a8clk.011330.jp" "383e3851cf.0115b2b799.com" "012024jhvjhkozekl.space" "0127c96640.com" "015fcec3c6.com" "019a6180a6.com" "01apple.com" "01b4075d6c.com" "01c70a2a06.com" "a8.01cloud.jp" "01counter.com" "1259e035ff.01f648fd79.com" "01jud3v55z.com" "dc.01menshealthblog.com" "securite.01net.com" "analytics.01networks.com" "02-apple-intelligence27.site" "ads.021.rs" "cname-aa.022022.net" "023hysj.com" "025g.top" "0265331.com" "img.0279.net" "027dir.com" "o.027eat.com" "0281.jp" "02asdf.com" "7452c605e9.02ea6adc36.com" "02ip.ru" "03180d2d.live" "031dfgh.com" "0322cfmtl.cc" "033ajy.xyz" "0342b40dd6.com" "oscnjc.035000.com" "03505ed0f4.com" "0351dvd.cn" "036e.cn" "03935e5357.com" "039nez.com" "03a17e7986.com" "03ab2qdorz.com" "03b5f525af.com" "03bdb617ed.com" "03c091d65a.com" "d55875dd70.03db0d5d14.com" "03e.info" "03e41ef81f.com" "03eea1b6dd.com" "03p.info" "a8cv.03plus.net" "pre.03porno.com" "04-f-bmf.com" "b004dc0e97.043213c072.com" "04424170.xyz" "044da016b3.com" "a8cv.04510.jp" "045fef1509.com" "04697a0ddea24c610de68229dadbc4390c37813647c2827bbc1f9041.com" "047e541483.com" "0497496f94.com" "d0ce8193a8.04b6001ba9.com" "fdaea0335d.04b73980ec.com" "04c8b396bf.com" "04ce9409dcb20470ed22f9967b00f1c1.net" "640186f930.04de88565b.com" "04e0d8fb0f.com" "04tips.com" "awklir.0506mall.com" "0512s.com" "9bff4acb16.051e57adf3.com" "052db.website" "053h94.com" "05420795.xyz" "05454674.xyz" "0571jjw.com" "appimg.0575bbs.com" "ffb7c5bd32.05951bf6a3.com" "pic.0597kk.com" "gg.0598yu.com" "059e71004b.com" "e275260174.05ae41c3fc.com" "05e11c9f6f.com" "05fbc08eac.com" "0676el9lskux.top" "06789.xyz" "06a21eff24.com" "06cffaae87.com" "f1.06ps.com" "070880.com" "07353.com" "0735sh.com" "f3dd3f65d2.0737522f52.com" "f6b458ade2.0740d1e3bf.com" "07421283.xyz" "0743j.com" "0755rc.com" "0760571ca9.com" "07634.com" "079301eaff0975107716716fd1cb0dcd.com" "07a1624bd7.com" "07aa269c0e76550c929640c170af557c7371753ba1b580236d7fa0a4.com" "bae5e9b94d.07aa35fee9.com" "07d0bc4a48.com" "07dy.cc" "1.07swz.com" "1919f4eccf.08031fef00.com" "08088.top" "080999.cn" "0816bvh.ru" "0820.com" "0833309e72.com" "0865a125ce.com" "087tqb.cn" "f237274791.0886c43482.com" "halo.088883.xyz" "08916fb8f8.com" "daae071ddb.08f12bcc45.com" "08f8f073.xyz" "08ro35delw.ru" "0909810.com" "0913u.com" "0916video.ru" "0926a687679d337e9d.com" "093093.jp" "9d0a3ce113.0935feb05f.com" "0940088.com" "0941.org" "09482aec5d.com" "095f2fc218.com" "09745951.xyz" "0981sldrltbr.live" "09b1fcc95e.com" "09zyy.com" "f97c68bed0.0a10a1bb7b.com" "82c39cef22.0a3036d0e7.com" "0a40c2b172.com" "0a65b6165b.com" "0438336acf.0a74314cd1.com" "0a8d87mlbcac.top" "0af2a962b0102942d9a7df351b20be55.com" "0al0zvucns.xyz" "0b0db57b5f.com" "ef34ee98f7.0b2d458c45.com" "05592cfcf1.0b383a4924.com" "0b57ha4k4zrd.top" "0b5bd8c4a8.com" "0b617f0769.com" "0b7741a902.com" "0b85c2f9bb.com" "8a9d20ab14.0b9b5eada8.com" "analytics.0bsnetwork.com" "0byv9mgbn0.com" "0c15ee8124.com" "0catch.com" "0cc29a3ac1.com" "0cdn.xyz" "0cf.io" "stats.0chris.com" "0cveatwx9onj.live" "0d076be0f4.com" "0d0c3ccf54.com" "56c30cd3c4.0d0e65883b.com" "0d4146061c.com" "01399322b4.0d4f63422a.com" "0d65577236.com" "0d76bd13e8.com" "5fd5fd02f8.0da9c10970.com" "0dax.com" "0df921ae7136d731950edfdd4876e3ee9759a5c2e63fafe0d43f45cf.com" "0didjsgheje.club" "0dwm.icu" "0e157d2cfa.com" "0e959bd265.com" "0eade9dd8d.com" "8a7d8912cb.0ef2473ad8.com" "0emm.com" "0er7pc8.xyz" "0f0b46245d.com" "6fb345e22a.0f3f317f1d.com" "0f461325bf56c3e1b9.com" "b00957f07f.0f7d2e9c5b.com" "ee768d2e27.0f9be2e1ef.com" "0faf13d8ae.com" "0fb.co" "0fb.ltd" "0fees.net" "0ffaf504b2.com" "0ffice395.net" "0geyfxqh2l.top" "0gw7e6s3wrao9y3q.pro" "0h3uds.com" "0i0i0i0.com" "0ijvby90.skin" "0j775d.cn" "0j91h.cyou" "0jcn4veha2oz.top" "0k7wod.cn" "0kal38g35ctc.top" "0l1201s548b2.top" "0m4.ru" "amateurs-cam.0my.net" "0n.click" "0n-line.tv" "0nedr1ve.com" "filter.a1.0network.com" "login.a1.0network.com" "xml.a1.0network.com" "xml-eu-v4.a1.0network.com" "xml-v4.a1.0network.com" "0ns32h.com" "ooo.0o0.ooo" "0op8kh.cn" "0oqt9i.cn" "0pdent.com" "0r17374.com" "0r3tyg.cn" "0redirc.com" "0redird.com" "0redire.com" "0rhagsowd21x.live" "0s09t235s.com" "0scan.com" "0sntp7dnrr.com" "0stats.com" "0storageatools0.xyz" "0sywjs4r1x.com" "0td6sdkfq.com" "0torrent.com" "0tq6ub.cn" "0u37vw5na9w0.top" "0utlook-microsoft.com" "0uvt8b.cn" "0uyt5b.cn" "0vijv8w.com" "stats.0www.org" "0x01n2ptpuz3.com" "metric.0x30c4.dev" "analytics.0xecho.com" "track.0xspotlight.xyz" "www12.0zz0.com" "www8.0zz0.com" "1.vg" "1-1ads.com" "1-2005-search.com" "a8cv.1-class.jp" "1-free-share-buttons.com" "1-office.info" "1-pregnant-sex.com" "1-remont.com" "100-flannelman.com" "oesonx.10000recipe.com" "awrgkd.1000farmacie.it" "1000lashes.com" "1000mercis.com" "1000russianwomen.com" "analytics.1000seen.de" "m1.ad.10010.com" "analytics.10010.com" "smartad.10010.com" "janzoz.1001pneus.fr" "matomo.1001services.net" "1002.men" "smetrics.1005freshradio.ca" "100669.com" "go.10086.cn" "sdc.10086.cn" "ads.100asians.com" "a9529eca57.100b57dcae.com" "ev.100calorias.com" "100conversions.com" "100dof.com" "100dollars-seo.com" "analytics.100dw.net" "100free.com" "webstats.100procent.com" "lan2.100second.com" "lans.100second.com" "100sexlinks.com" "100widgets.com" "100womenwhocareoakville.com" "1.1010pic.com" "smetrics.1011bigfm.com" "1017.cn" "1017facai8.xyz" "101bentonstreet.com" "101billion.com" "101com.com" "101flag.ru" "101freehost.com" "101m3.com" "0.101tubeporn.com" "analytics.102449331.com" "1024mzs.pw" "1028images.com" "pre.102porno.club" "pre.102porno.net" "103092804.com" "smetrics.1031freshradio.ca" "metric.1035thearrow.com" "103bees.com" "c4038bd4ca.103dc14b45.com" "smetrics.1043freshradio.ca" "smetrics.1045freshradio.ca" "10523745.xyz" "105app.com" "oms.1067rock.ca" "10753990.xyz" "smetrics.1075daverocks.com" "107e9a08a8.com" "108.com" "nafmxc.1083.fr" "108shot.com" "108topreviews.com" "1090pjopm.de" "af043ebde4.10a03eb82c.com" "2c33f8d1d4.10b7647bbf.com" "10bet.com" "analytics.10bshop.com" "10c26a1dd6.com" "10cd.ru" "10desires.com" "plausible.10e-9.dev" "t.10er-tagesticket.de" "analytics.10fastfingers.com" "umami.10fastfingers.com" "analytics.10g8.com" "analytics.10gb.es" "tracking.10gb.vn" "10gbfreehost.com" "analytics.10hunter.ca" "sadbmetrics.10knocturnagijon.es" "metric.10ktf.com" "analytics.10life.com" "10nvejhblhha.com" "10q6e9ne5.de" "10sn95to9.de" "web-ads.10sq.net" "ywrcqa.10tv.com" "10un.jp" "10vekatu.jp" "metrics.10web.io" "10xtask.com" "google.com")

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
