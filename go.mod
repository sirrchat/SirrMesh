module github.com/mail-chat-chain/sirrmeshd

go 1.24

toolchain go1.24.4

require (
	blitiri.com.ar/go/spf v1.5.1
	github.com/GehirnInc/crypt v0.0.0-20230320061759-8cc1b52080c5
	github.com/c0va23/go-proxyprotocol v0.9.1
	github.com/caddyserver/certmagic v0.21.7
	github.com/emersion/go-imap v1.2.2-0.20220928192137-6fac715be9cf
	github.com/emersion/go-imap-compress v0.0.0-20201103190257-14809af1d1b9
	github.com/emersion/go-imap-sortthread v1.2.0
	github.com/emersion/go-message v0.18.2
	github.com/emersion/go-milter v0.4.1
	github.com/emersion/go-msgauth v0.7.0
	github.com/emersion/go-sasl v0.0.0-20241020182733-b788ff22d5a6
	github.com/emersion/go-smtp v0.24.0
	github.com/ethereum/go-ethereum v1.15.11
	github.com/foxcpp/go-dovecot-sasl v0.0.0-20200522223722-c4699d7a24bf
	github.com/foxcpp/go-imap-backend-tests v0.0.0-20220105184719-e80aa29a5e16
	github.com/foxcpp/go-imap-i18nlevel v0.0.0-20200208001533-d6ec88553005
	github.com/foxcpp/go-imap-mess v0.0.0-20230108134257-b7ec3a649613
	github.com/foxcpp/go-imap-namespace v0.0.0-20200802091432-08496dd8e0ed
	github.com/foxcpp/go-imap-sql v0.5.1-0.20250124140007-8da5567429d5
	github.com/foxcpp/go-mockdns v1.1.0
	github.com/foxcpp/go-mtasts v0.0.0-20240130093538-1438da2e5932
	github.com/go-ldap/ldap/v3 v3.4.11
	github.com/go-sql-driver/mysql v1.7.1
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-hclog v1.6.3
	github.com/johannesboyne/gofakes3 v0.0.0-20250825084532-6555d310c473
	github.com/lib/pq v1.10.9
	github.com/libdns/acmedns v0.2.0
	github.com/libdns/alidns v1.0.3
	github.com/libdns/cloudflare v0.1.1
	github.com/libdns/digitalocean v0.0.0-20230728223659-4f9064657aea
	github.com/libdns/gandi v1.0.3
	github.com/libdns/gcore v0.0.0-20250127070537-4a9d185c9d20
	github.com/libdns/googleclouddns v1.1.0
	github.com/libdns/hetzner v0.0.1
	github.com/libdns/leaseweb v0.4.0
	github.com/libdns/libdns v0.2.2
	github.com/libdns/metaname v0.3.0
	github.com/libdns/namecheap v0.0.0-20211109042440-fc7440785c8e
	github.com/libdns/namedotcom v0.3.3
	github.com/libdns/rfc2136 v0.1.1
	github.com/libdns/route53 v1.5.1
	github.com/libdns/vultr v1.0.0
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/miekg/dns v1.1.64
	github.com/minio/minio-go/v7 v7.0.95
	github.com/netauth/netauth v0.6.4
	github.com/prometheus/client_golang v1.22.0
	github.com/spf13/cobra v1.9.1
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.41.0
	golang.org/x/net v0.43.0
	golang.org/x/sync v0.16.0
	golang.org/x/term v0.34.0
	golang.org/x/text v0.28.0
	modernc.org/sqlite v1.18.1
)

require (
	cloud.google.com/go v0.121.0 // indirect
	cloud.google.com/go/auth v0.16.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/DataDog/zstd v1.5.7 // indirect
	github.com/G-Core/gcore-dns-sdk-go v0.3.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.36.3 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.14 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.67 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/route53 v1.42.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.19 // indirect
	github.com/aws/smithy-go v1.22.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.24.0 // indirect
	github.com/caddyserver/zerossl v0.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/errors v1.12.0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20241215232642-bb51bb14a506 // indirect
	github.com/cockroachdb/redact v1.1.6 // indirect
	github.com/consensys/gnark-crypto v0.18.0 // indirect
	github.com/crate-crypto/go-eth-kzg v1.3.0 // indirect
	github.com/crate-crypto/go-ipa v0.0.0-20240724233137-53bbb0ceb27a // indirect
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/digitalocean/godo v1.148.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/ethereum/c-kzg-4844/v2 v2.1.0 // indirect
	github.com/ethereum/go-verkle v0.2.2 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/getsentry/sentry-go v0.32.0 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.8-0.20250403174932-29230038a667 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/holiman/uint256 v1.3.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mholt/acmez/v3 v3.1.2 // indirect
	github.com/minio/crc64nvme v1.0.2 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/netauth/protocol v0.0.0-20210918062754-7fee492ffcbd // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.63.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/ryszard/goskiplist v0.0.0-20150312221310-2dfbae5fcf46 // indirect
	github.com/sagikazarmark/locafero v0.10.0 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.14.0 // indirect
	github.com/spf13/cast v1.9.2 // indirect
	github.com/spf13/pflag v1.0.7 // indirect
	github.com/spf13/viper v1.20.1 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/supranational/blst v0.3.14 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/tinylib/msgp v1.3.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/vultr/govultr/v3 v3.0.2 // indirect
	github.com/zeebo/assert v1.3.0 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0 // indirect
	go.opentelemetry.io/otel v1.36.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/trace v1.36.0 // indirect
	go.shabbyrobe.org/gocovmerge v0.0.0-20230507111327-fa4f82cfbf4d // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap/exp v0.3.0 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	google.golang.org/api v0.233.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250811230008-5f3141c8851a // indirect
	google.golang.org/grpc v1.74.2 // indirect
	google.golang.org/protobuf v1.36.7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools v2.2.0+incompatible // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.36.3 // indirect
	modernc.org/ccgo/v3 v3.16.9 // indirect
	modernc.org/libc v1.17.1 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.2.1 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.0.0 // indirect
)

replace (
	github.com/emersion/go-imap => github.com/foxcpp/go-imap v1.0.0-beta.1.0.20220623182312-df940c324887
	github.com/emersion/go-smtp => github.com/foxcpp/go-smtp v1.21.4-0.20250124171104-c8519ae4fb23
	// use Cosmos geth fork for ethereum compatibility
	github.com/ethereum/go-ethereum => github.com/cosmos/go-ethereum v0.0.0-20250806193535-2fc7571efa91
	github.com/libdns/gandi => github.com/foxcpp/libdns-gandi v1.0.4-0.20240127130558-4782f9d5ce3e
	// replace broken goleveldb
	github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
)
