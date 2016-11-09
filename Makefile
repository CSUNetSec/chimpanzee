default: pb_pull pb_compile_common pb_compile_proddle pb_compile_netbrane

pb_dir_init:
	mkdir src/pb; \
	echo "pub mod common;" >> src/pb/mod.rs; \
	echo "pub mod proddle;" >> src/pb/mod.rs; \
	echo "pub mod bgp;" >> src/pb/mod.rs; \
	echo "pub mod netbrane;" >> src/pb/mod.rs;

pb_init: pb_dir_init
	git init protobuf; \
	cd protobuf; \
	git remote add -f origin https://github.com/CSUNetSec/netsec-protobufs.git; \
	git config core.sparseCheckout true; \
	echo "common" >> .git/info/sparse-checkout; \
	echo "proddle" >> .git/info/sparse-checkout; \
	echo "protocol" >> .git/info/sparse-checkout; \
	echo "netbrane" >> .git/info/sparse-checkout;

pb_pull:
	cd protobuf; \
	git pull origin master;

pb_compile_common:
	protoc --rust_out=src/pb protobuf/common/common.proto;

pb_compile_proddle:
	protoc --rust_out=src/pb protobuf/proddle/proddle.proto;

pb_compile_bgp:
	sed 's#github.com/CSUNetSec/netsec-protobufs/common/common.proto#common.proto#' protobuf/protocol/bgp/bgp.proto > /tmp/bgp.proto; \
	protoc -I /tmp -I protobuf/common --rust_out=src/pb /tmp/bgp.proto;

pb_compile_netbrane: pb_compile_bgp
	sed 's#github.com/CSUNetSec/netsec-protobufs/common/common.proto#common.proto#' protobuf/netbrane/netbrane.proto > /tmp/netbrane.proto.common; \
	sed 's#github.com/CSUNetSec/netsec-protobufs/proddle/proddle.proto#proddle.proto#' /tmp/netbrane.proto.common > /tmp/netbrane.proto.proddle; \
	sed 's#github.com/CSUNetSec/netsec-protobufs/protocol/bgp/bgp.proto#bgp.proto#' /tmp/netbrane.proto.proddle > /tmp/netbrane.proto; \
	protoc -I /tmp -I protobuf/common/ -I protobuf/proddle/ -I protobuf/protocol/bgp --rust_out=src/pb/ /tmp/netbrane.proto; \
	rm /tmp/netbrane*; \
	rm /tmp/*.proto
