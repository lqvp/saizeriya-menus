PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "categories" (
	"id"	INTEGER NOT NULL UNIQUE,
	"name"	TEXT NOT NULL UNIQUE,
	"name_en"	TEXT NOT NULL UNIQUE,
	"name_zh"	TEXT NOT NULL UNIQUE,
	PRIMARY KEY("id" AUTOINCREMENT)
);
INSERT INTO categories VALUES(1,'グランド','Grand Menu','主');
INSERT INTO categories VALUES(2,'ランチ','Lunch','午餐');
INSERT INTO categories VALUES(3,'モーニング','Morning','早餐');
INSERT INTO categories VALUES(4,'キッズ','Kids','孩子们');
INSERT INTO categories VALUES(5,'テイクアウト','Take-out','外带');
INSERT INTO categories VALUES(6,'季節限定','Seasonal','季节限定');
CREATE TABLE IF NOT EXISTS "genres" (
	"id"	INTEGER NOT NULL UNIQUE,
	"name"	TEXT NOT NULL UNIQUE, pre_id TEXT,
	PRIMARY KEY("id")
);
INSERT INTO genres VALUES(12,'サラダ','SA');
INSERT INTO genres VALUES(13,'スープ','SU');
INSERT INTO genres VALUES(14,'前菜・おつまみ','AA');
INSERT INTO genres VALUES(21,'ドリア&グラタン','DG');
INSERT INTO genres VALUES(22,'ピザ','PZ');
INSERT INTO genres VALUES(23,'パスタ','PA');
INSERT INTO genres VALUES(24,'ハンバーグ','MT');
INSERT INTO genres VALUES(31,'ライス・パン','RP');
INSERT INTO genres VALUES(32,'デザート (すぐに)','DE');
INSERT INTO genres VALUES(33,'ビール・サワー・ノンアルコール','BR');
INSERT INTO genres VALUES(34,'ワイン','WN');
INSERT INTO genres VALUES(39,'デザート (あとで)','DE');
INSERT INTO genres VALUES(43,'トッピング','TP');
INSERT INTO genres VALUES(51,'ドリンクバー','DB');
INSERT INTO genres VALUES(53,'ドレッシング・オリーブオイル',NULL);
CREATE TABLE IF NOT EXISTS "icons" (
	"id"	INTEGER NOT NULL UNIQUE,
	"icon"	TEXT NOT NULL UNIQUE,
	PRIMARY KEY("id" AUTOINCREMENT)
);
INSERT INTO icons VALUES(1,'🥗');
INSERT INTO icons VALUES(2,'🥣');
INSERT INTO icons VALUES(3,'🍗');
INSERT INTO icons VALUES(4,'🍢');
INSERT INTO icons VALUES(5,'🥬');
INSERT INTO icons VALUES(6,'🍤');
INSERT INTO icons VALUES(7,'🥘');
INSERT INTO icons VALUES(8,'🌭');
INSERT INTO icons VALUES(9,'🐔');
INSERT INTO icons VALUES(10,'🦪');
INSERT INTO icons VALUES(11,'🥕');
INSERT INTO icons VALUES(12,'🍅');
INSERT INTO icons VALUES(13,'🥔');
INSERT INTO icons VALUES(14,'🍖');
INSERT INTO icons VALUES(15,'🧀');
INSERT INTO icons VALUES(16,'🫛');
INSERT INTO icons VALUES(17,'🍕');
INSERT INTO icons VALUES(18,'🍝');
INSERT INTO icons VALUES(19,'🍚');
INSERT INTO icons VALUES(20,'🍞');
INSERT INTO icons VALUES(21,'🥖');
INSERT INTO icons VALUES(22,'🍨');
INSERT INTO icons VALUES(23,'🍮');
INSERT INTO icons VALUES(24,'🍰');
INSERT INTO icons VALUES(25,'🍺');
INSERT INTO icons VALUES(26,'🍷');
INSERT INTO icons VALUES(27,'🍾');
INSERT INTO icons VALUES(28,'🧁');
INSERT INTO icons VALUES(29,'🥚');
INSERT INTO icons VALUES(30,'🥤');
INSERT INTO icons VALUES(31,'🧂');
CREATE TABLE IF NOT EXISTS "menus" (
    id INTEGER NOT NULL UNIQUE,
    name TEXT NOT NULL,
    name_en TEXT,
    name_zh TEXT,
    price INTEGER NOT NULL,
    price_with_tax INTEGER NOT NULL,
    calorie INTEGER,
    salt INTEGER,
    category_id INTEGER,
    genre_id INTEGER,
    is_alcohol INTEGER DEFAULT false,
    icon_id INTEGER,
    pre_id TEXT,
    PRIMARY KEY(id),
    FOREIGN KEY(category_id) REFERENCES categories(id),
    FOREIGN KEY(genre_id) REFERENCES genres(id),
    FOREIGN KEY(icon_id) REFERENCES icons(id)
);
INSERT INTO menus VALUES(1202,'小エビのサラダ','Shrimp Salad','小甜虾沙拉',319,350,192,1.5,1,12,0,6,'SA02');
INSERT INTO menus VALUES(1205,'わかめのサラダ','Seaweed Salad','海藻沙拉',319,350,174,2.9,1,12,0,1,'SA05');
INSERT INTO menus VALUES(1207,'モッツァレラのサラダ','Green Salad with Buffalo Mozzarella Cheese','水牛马苏里拉芝士蔬菜沙拉',364,400,243,0.6,1,12,0,1,'SA07');
INSERT INTO menus VALUES(1208,'グリーンサラダ','Green Salad','绿色的沙拉',319,350,166,0.7,1,12,0,1,'SA08');
INSERT INTO menus VALUES(1209,'チキンのサラダ','Chicken Salad','鸡肉沙拉',319,350,224,1.3,1,12,0,1,'SA03');
INSERT INTO menus VALUES(1301,'コーンクリームスープ','Creamy Corn Soup','奶油玉米浓汤',137,150,154,1,1,13,0,2,'SU01');
INSERT INTO menus VALUES(1305,'田舎風ミネストローネ','Country-style Minestrone','田园风蔬菜汤',273,300,227,3,1,13,0,2,'SU05');
INSERT INTO menus VALUES(1307,'たまねぎのズッパ','Onion soup, Zuppa','洋葱面包汤',273,300,228,2,1,13,0,2,'SU07');
INSERT INTO menus VALUES(1401,'辛味チキン','Spicy Grilled Chicken','香辣鸡翅',273,300,295,1.8,1,14,0,3,'AA01');
INSERT INTO menus VALUES(1402,'アロスティチーニ（ラムの串焼き）','Arrosticini (Lamb skewers)','意式羊肉串（烤羊肉串）',364,400,224,0.6,1,14,0,4,'AA02');
INSERT INTO menus VALUES(1403,'ほうれん草のソテー','Sauteed Spinach（Salt Oil Sauce）','炖煮菠菜（盐油酱）',182,200,223,1.2,1,14,0,5,'AA03');
INSERT INTO menus VALUES(1404,'ポップコーンシュリンプ','Popcorn Shrimp','炸小甜虾',273,300,234,1.2,1,14,0,6,'AA04');
INSERT INTO menus VALUES(1405,'エスカルゴのオーブン焼き','Grilled Escargots','蒜香焗蜗牛',364,400,220,1.3,1,14,0,7,'AA05');
INSERT INTO menus VALUES(1406,'小エビのカクテル','Shrimp with "SAIZERIYA" Dressing','虾佐“SAIZERIYA”酱汁',255,280,126,2.3,1,14,0,6,'AA06');
INSERT INTO menus VALUES(1407,'チョリソー（辛味ソーセージ）','Grilled Spicy Sausages','口利左香肠（辣猪肉香肠）',364,400,425,2.9,1,14,0,8,'AA07');
INSERT INTO menus VALUES(1408,'蒸し鶏の香味ソース','Chicken with Flavored Vegetable Sauce（Limited to some stores）','鸡肉蔬菜酱（限部分店鋪）',255,280,192,2.4,1,14,0,9,'AA08');
INSERT INTO menus VALUES(1410,'ムール貝のガーリック焼き','Grilled Mussels with Vegetable Salsa','蒜香烤蛤蜊',364,400,170,1.4,1,14,0,10,'AA10');
INSERT INTO menus VALUES(1411,'アスパラガスの温サラダ','Asparagus with  Cheese ','芦笋配山核桃奶酪',273,300,80,0.7,1,14,0,1,'AA09');
INSERT INTO menus VALUES(1413,'キャロットラペ','Refreshing Carrot Salad','爽口胡萝卜沙拉',182,200,93,0.9,1,14,0,11,'AA13');
INSERT INTO menus VALUES(1414,'モッツァレラトマト','Buffalo Mozzarella Cheese &  Tomato','布法罗马苏里拉奶酪和番茄',391,430,255,0.2,1,14,0,12,'AA14');
INSERT INTO menus VALUES(1415,'カリッとポテト','Crispy Potatoes','酥烤马铃薯',255,280,389,2.1,1,14,0,13,'AA15');
INSERT INTO menus VALUES(1416,'ポテトのグリル','Hashed Potato',NULL,273,300,307,1.6,1,14,0,13,'');
INSERT INTO menus VALUES(1417,'バッファローモッツァレラのカプレーゼ','Buffalo Mozzarella Cheese &  Tomato',NULL,391,430,254,0.4,1,14,0,12,'');
INSERT INTO menus VALUES(1422,'ハモン・セラーノ','Jamon Serrano','火腿塞拉诺',291,320,39,0.9,1,14,0,14,'AA22');
INSERT INTO menus VALUES(1423,'生ハムとバッファローモッツァレラの盛合せ','Jamon Serrano &  Buffalo Mozzarella Cheese','火腿塞拉诺和水牛马苏里拉奶酪',455,500,146,1,1,14,0,15,NULL);
INSERT INTO menus VALUES(1425,'柔らか青豆の温サラダ','Green Peas with  Cheese & Egg','鲜嫩青豆加佩科里诺干酪温色拉',182,200,216,1.4,1,14,0,16,'AA06');
INSERT INTO menus VALUES(1435,'スイートコーン','Sweet Corn','',182,200,NULL,NULL,1,14,0,1,'');
INSERT INTO menus VALUES(1452,'アロスティチーニ（ラムの串焼き）(Wサイズ)','Arrosticini (Lamb skewers)','意式羊肉串（烤羊肉串）',728,800,NULL,NULL,1,14,0,4,'AA52');
INSERT INTO menus VALUES(2101,'ミラノ風ドリア','Meat Sauce Doria','米兰风肉酱多利亚饭',273,300,521,2.5,1,21,0,7,'DG01');
INSERT INTO menus VALUES(2103,'半熟卵のミラノ風ドリア','Meat Sauce & Coddled Egg Doria','半熟蛋配米兰风多利亚肉酱饭',319,350,604,2.7,1,21,0,7,'DG03');
INSERT INTO menus VALUES(2106,'タラコとエビのドリア','Doria with “TARAKO”Sauce & Shrimp','鳕鱼籽虾焗饭',364,400,578,2.8,1,21,0,7,'DG06');
INSERT INTO menus VALUES(2108,'焼チーズ ミラノ風ドリア','Doria with Meat Sauce & Cheese','多利亚肉酱和奶酪',319,350,652,2.9,1,21,0,7,'DG08');
INSERT INTO menus VALUES(2109,'エビとタラコのクリームグラタン（全粒粉）','Gratin with “TARAKO”Sauce & Shrimp','TARAKO醬蝦焗烤',391,400,481,2.1,1,21,0,7,NULL);
INSERT INTO menus VALUES(2110,'タラコとポップコーンシュリンプのドリア',NULL,NULL,364,400,617,2.8,1,21,0,7,NULL);
INSERT INTO menus VALUES(2115,'ポップコーンシュリンプとタラコのクリームグラタン（全粒粉）',NULL,NULL,391,430,591,2.6,1,21,0,7,NULL);
INSERT INTO menus VALUES(2203,'バッファローモッツァレラのマルゲリータピザ','Pizza with  Buffalo Mozzarella Cheese','水牛马苏里拉披萨',364,400,544,2.4,1,22,0,17,'PZ02');
INSERT INTO menus VALUES(2204,'野菜ときのこのピザ','Pizza with Mushrooms & Vegetable Salsa','蔬菜蘑菇比萨',364,400,593,3.2,1,22,0,17,'PZ04');
INSERT INTO menus VALUES(2206,'たっぷりコーンのピザ','Corn Pizza','足量甜玉米比萨',364,400,624,2.6,1,22,0,17,'PZ06');
INSERT INTO menus VALUES(2208,'ソーセージピザ','Sausage Pizza','香肠比萨',364,400,710,4.1,1,22,0,17,'PZ08');
INSERT INTO menus VALUES(2301,'タラコソースシシリー風',NULL,NULL,364,400,538,2.2,1,23,0,18,NULL);
INSERT INTO menus VALUES(2303,'ペペロンチーノ','Spaghetti Peperoncino','香辣意面',273,300,582,1.9,1,23,0,18,'PA03');
INSERT INTO menus VALUES(2304,'パルマ風スパゲティ','Spaghetti with tomato＆pancetta','番茄培根意大利面',364,400,753,3.1,1,23,0,18,'PA04');
INSERT INTO menus VALUES(2305,'カルボナーラ','Spaghetti Carbonara','奶油培根意面',455,500,799,3.1,1,23,0,18,'PA05');
INSERT INTO menus VALUES(2306,'ミートソースボロニア風','Spaghetti Meat Sauce','意大利肉酱面',364,400,604,3.6,1,23,0,18,'PA02');
INSERT INTO menus VALUES(2310,'スープ入り塩味ボンゴレ','Spaghetti"VONGOLE"','意大利面“VONGOLE”',455,500,872,2.5,1,23,0,18,'PA10');
INSERT INTO menus VALUES(2316,'半熟卵のミートソースボロニア風','Spaghetti Meat Sauce with Coddled Egg','半熟蛋配意大利肉酱面',410,450,673,3.8,1,23,0,18,'PA12');
INSERT INTO menus VALUES(2317,'半熟卵のぺペロンチーノ','Spaghetti Peperoncino with soft-boiled egg','半熟蛋配香辣意面',319,350,665,2.1,1,23,0,18,'PA17');
INSERT INTO menus VALUES(2320,'小エビのタラコソース','Spaghetti “TARAKO” Sauce with Shrimp','意大利面“TARAKO”虾酱',491,540,594,2.8,1,23,0,18,'PA20');
INSERT INTO menus VALUES(2321,'きのことほうれん草のクリームスパゲッティ','Spaghetti with spinach and mushroom cream','菠菜蘑菇奶油意大利面',546,600,856,3.7,1,23,0,18,NULL);
INSERT INTO menus VALUES(2325,'ペンネアラビアータ（全粒粉）','Whole wheat penne arrabbiata','全麥通心粉',391,400,444,2.8,1,23,0,18,'PA25');
INSERT INTO menus VALUES(2328,'イカの墨入りセピアソース','Spaghetti al Nero di Seppia','墨鱼汁海鲜意面',455,500,603,1.8,1,23,0,18,NULL);
INSERT INTO menus VALUES(2402,'若鶏のディアボラ風','Grilled Chicken with Chopped Vegetable Sauce','意式恶魔风嫩鸡排',455,500,673,2.8,1,24,0,9,'MT02');
INSERT INTO menus VALUES(2403,'イタリアンハンバーグ','Italian Hamburg Steak','意大利风味汉堡牛排',455,500,670,3,1,24,0,14,'MT03');
INSERT INTO menus VALUES(2404,'柔らかチキンのチーズ焼き','Grilled Chicken with Cheese','意大利风芝士鸡排',455,500,725,2.8,1,24,0,9,'MT04');
INSERT INTO menus VALUES(2406,'ハンバーグステーキ','Hamburg Steak(Demi-Glace Sauce)','汉堡牛排(半冰酱)',364,400,571,2.5,1,24,0,14,'MT01');
INSERT INTO menus VALUES(2407,'ディアボラ風ハンバーグ','Hamburg Steak with Chopped Vegetable Sauce','意式恶魔风汉堡牛排',455,500,618,3.1,1,24,0,14,'MT07');
INSERT INTO menus VALUES(2413,'ラム（仔羊）と野菜のグリル','Grilled Lamb ＆ Vegetables','烤羊肉＆蔬菜',791,870,371,2.4,1,24,0,14,'MT13');
INSERT INTO menus VALUES(2417,'ラムのランプステーキ','Lamb Rump Steak',NULL,991,1090,357,2.2,1,24,0,14,NULL);
INSERT INTO menus VALUES(2418,'ミックスグリル','Mixed Grill','混合烧烤',591,650,779,3.6,1,24,0,14,NULL);
INSERT INTO menus VALUES(2419,'ビーフステーキ(ガルムソース)','Beef Steak','牛排',991,1090,374,2.5,1,24,0,14,NULL);
INSERT INTO menus VALUES(3101,'ライス','Rice','米饭',137,150,303,NULL,1,31,0,19,'RP01');
INSERT INTO menus VALUES(3102,'ラージライス','Large Rice','米饭',182,200,303,NULL,1,31,0,19,'RP02');
INSERT INTO menus VALUES(3103,'スモールライス','Small Rice','米饭',91,100,303,NULL,1,31,0,19,'RP03');
INSERT INTO menus VALUES(3104,'シナモンプチフォッカ','Petit Focaccia with Cinnamon Sugar','肉桂糖迷你香草面包',182,200,246,0.8,1,31,0,20,'RP04');
INSERT INTO menus VALUES(3106,'プチフォッカ','Petit Focaccias','迷你香草面包',137,150,214,0.8,1,31,0,20,'RP06');
INSERT INTO menus VALUES(3108,'ミニフィセル','Mini Ficelle','迷你法式面包',137,150,189,1,1,31,0,21,'RP08');
INSERT INTO menus VALUES(3109,'ガーリックトースト','Garlic Toast','蒜香烤面包',182,200,245,1.1,1,31,0,21,'RP09');
INSERT INTO menus VALUES(3110,'フォッカチオ',NULL,NULL,137,150,241,0.6,1,31,0,20,NULL);
INSERT INTO menus VALUES(3111,'ガーリックフォッカチオ',NULL,NULL,182,200,297,0.8,1,31,0,20,NULL);
INSERT INTO menus VALUES(3112,'シナモンフォッカチオ',NULL,NULL,182,200,273,0.6,1,31,0,20,NULL);
INSERT INTO menus VALUES(3201,'ティラミス　クラシコ','Tiramisu','提拉米苏经典系列',273,300,229,0.1,1,32,0,28,'DE01');
INSERT INTO menus VALUES(3204,'ジェラート＆シナモンプチフォッカ','Petit Focaccias with Cinamon Sugar & Italian Gelato','肉桂糖迷你香草面包配意大利冰激淋',410,450,372,0.9,1,32,0,22,'DE04');
INSERT INTO menus VALUES(3205,'ミルクジェラート','Italian Gelato','意大利冰激淋',228,250,121,0.1,1,32,0,22,'DE05');
INSERT INTO menus VALUES(3206,'イタリアンプリン','Italian Cream Caramel','意大利布丁',228,250,216,0.1,1,32,0,23,'DE06');
INSERT INTO menus VALUES(3207,'チョコレートケーキ','Chocolate Cake','巧克力蛋糕',273,300,166,0.1,1,32,0,24,'DE07');
INSERT INTO menus VALUES(3212,'プリンとティラミス　クラシコの盛合せ','Tiramisu & Italian Cream Caramel','布丁配提拉米苏',455,500,445,0.2,1,32,0,23,'DE12');
INSERT INTO menus VALUES(3213,'トリフアイスクリーム','Ice Cream Truffle','松露冰激淋',319,350,164,0.1,1,32,0,22,'DE13');
INSERT INTO menus VALUES(3214,'ジェラート＆シナモンフォッカチオ','Gelato & Cinnamon Focaccia','冰激淋配肉桂佛卡夏',410,450,393,0.7,1,32,0,22,NULL);
INSERT INTO menus VALUES(3215,'コーヒーゼリー＆ミルクジェラート','Coffee Jelly with Italian Gelato','意大利冰激淋配咖啡果冻',319,350,165,0.1,1,32,0,22,'DE15');
INSERT INTO menus VALUES(3301,'生ビール キリン一番搾り　ジョッキ',NULL,NULL,364,400,NULL,NULL,1,33,1,25,'BR01');
INSERT INTO menus VALUES(3302,'生ビール キリン一番搾り　グラスビール',NULL,NULL,273,300,NULL,NULL,1,33,1,25,'BR02');
INSERT INTO menus VALUES(3303,'ノンアルコール アサヒドライゼロ',NULL,NULL,228,250,NULL,NULL,1,33,0,25,NULL);
INSERT INTO menus VALUES(3304,'サワー キリン氷結シチリア産レモン',NULL,NULL,319,350,NULL,NULL,1,33,1,25,'BR04');
INSERT INTO menus VALUES(3306,'グラッパ (30ml)',NULL,NULL,273,300,NULL,NULL,1,33,1,25,NULL);
INSERT INTO menus VALUES(3401,'グラスワイン (120ml)　赤',NULL,NULL,91,100,NULL,NULL,1,34,1,26,'WN01');
INSERT INTO menus VALUES(3402,'グラスワイン (120ml)　白',NULL,NULL,91,100,NULL,NULL,1,34,1,26,'WN02');
INSERT INTO menus VALUES(3403,'デカンタ (250ml)　赤',NULL,NULL,182,200,NULL,NULL,1,34,1,26,'WN03');
INSERT INTO menus VALUES(3404,'デカンタ (250ml)　白',NULL,NULL,182,200,NULL,NULL,1,34,1,26,'WN04');
INSERT INTO menus VALUES(3405,'デカンタ (500ml)　赤',NULL,NULL,364,400,NULL,NULL,1,34,1,26,'WN05');
INSERT INTO menus VALUES(3406,'デカンタ (500ml)　白',NULL,NULL,364,400,NULL,NULL,1,34,1,26,'WN06');
INSERT INTO menus VALUES(3407,'マグナム (1500ml)　赤',NULL,NULL,1000,1100,NULL,NULL,1,34,1,27,'WN07');
INSERT INTO menus VALUES(3408,'マグナム (1500ml)　白',NULL,NULL,1000,1100,NULL,NULL,1,34,1,27,'WN08');
INSERT INTO menus VALUES(3412,'ランブルスコロゼ【(ロゼ・発泡)甘口】',NULL,NULL,1000,1100,NULL,NULL,1,34,1,27,'WN12');
INSERT INTO menus VALUES(3413,'ドンラファエロ【(白・発泡)辛口】',NULL,NULL,1000,1100,NULL,NULL,1,34,1,27,'WN13');
INSERT INTO menus VALUES(3414,'ランブルスコセッコ【(赤・発泡)辛口】',NULL,NULL,1000,1100,NULL,NULL,1,34,1,27,'WN14');
INSERT INTO menus VALUES(3415,'ベルデッキオ【(白)辛口】',NULL,NULL,1000,1100,NULL,NULL,1,34,1,27,'WN15');
INSERT INTO menus VALUES(3416,'キャンティ【(赤)辛口】',NULL,NULL,1000,1100,NULL,NULL,1,34,1,27,'WN16');
INSERT INTO menus VALUES(3419,'キャンティ ルフィナ リゼルバ【(赤)辛口】',NULL,NULL,2000,2200,NULL,NULL,1,34,1,27,'WN19');
INSERT INTO menus VALUES(3901,'ティラミス　クラシコ','Tiramisu','提拉米苏经典系列',273,300,229,0.1,1,39,0,28,'DE01');
INSERT INTO menus VALUES(3904,'ジェラート＆シナモンプチフォッカ','Petit Focaccias with Cinamon Sugar & Italian Gelato','肉桂糖迷你香草面包配意大利冰激淋',410,450,372,0.9,1,39,0,22,'DE04');
INSERT INTO menus VALUES(3905,'ミルクジェラート','Italian Gelato','意大利冰激淋',228,250,121,0.1,1,39,0,22,'DE05');
INSERT INTO menus VALUES(3906,'イタリアンプリン','Italian Cream Caramel','意大利布丁',228,250,216,0.1,1,39,0,23,'DE06');
INSERT INTO menus VALUES(3907,'チョコレートケーキ','Chocolate Cake','巧克力蛋糕',273,300,166,0.1,1,39,0,24,'DE07');
INSERT INTO menus VALUES(3912,'プリンとティラミス　クラシコの盛合せ','Tiramisu & Italian Cream Caramel','布丁配提拉米苏',455,500,445,0.2,1,39,0,23,'DE12');
INSERT INTO menus VALUES(3913,'トリフアイスクリーム','Ice Cream Truffle','松露冰激淋',319,350,164,0.1,1,39,0,22,'DE13');
INSERT INTO menus VALUES(3914,'ジェラート＆シナモンフォッカチオ','Gelato & Cinnamon Focaccia','冰激淋配肉桂佛卡夏',410,450,393,0.7,1,32,0,22,NULL);
INSERT INTO menus VALUES(3915,'コーヒーゼリー＆ミルクジェラート','Coffee Jelly with Italian Gelato','意大利冰激淋配咖啡果冻',319,350,165,0.1,1,39,0,22,'DE15');
INSERT INTO menus VALUES(4301,'トッピング半熟卵','Coddled Egg for Topping','加半熟蛋',46,50,83,0.2,1,43,0,29,'TP01');
INSERT INTO menus VALUES(4304,'トッピング野菜ソース','Vegetable Paste','蔬菜酱',91,100,72,0.3,1,43,0,1,'TP04');
INSERT INTO menus VALUES(4307,'トッピング粉チーズ（グランモラビア）','Table Cheese','加芝士粉（摩拉維並干酪）',91,100,91,0.6,1,43,0,15,'TP07');
INSERT INTO menus VALUES(5101,'セットドリンクバー','Drink Bar (When Ordered with Food)','畅饮吧（畅饮吧套餐）',182,200,NULL,NULL,1,51,0,30,'DB01');
INSERT INTO menus VALUES(5102,'キッズドリンクバー','Drink Bar (Kids)','畅饮吧（儿童）',91,100,NULL,NULL,4,51,0,30,'DB02');
INSERT INTO menus VALUES(5103,'ドリンクバー単品','Drink Bar (Only)','畅饮吧（單品）',273,300,NULL,NULL,1,51,0,30,'DB03');
INSERT INTO menus VALUES(5305,'サイゼリヤ　ドレッシング (500ml)',NULL,NULL,463,500,NULL,NULL,1,53,0,31,NULL);
INSERT INTO menus VALUES(5306,'エクストラ・バージン・オリーブオイル (500ml)',NULL,NULL,1112,1200,NULL,NULL,1,53,0,31,NULL);
INSERT INTO menus VALUES(5601,'ハンバーグモーニング','Hamburg Steak Morning','汉堡牛排早餐',909,1000,868,6.5,3,24,0,14,NULL);
INSERT INTO menus VALUES(5602,'タラコスパゲッティモーニング','Tarako Spaghetti Morning','鳕鱼籽意面早餐',909,1000,805,6.2,3,23,0,18,NULL);
INSERT INTO menus VALUES(5603,'フォッカチオモーニング','Focaccia Morning','佛卡夏早餐',909,1000,897,6.8,3,31,0,20,NULL);
INSERT INTO menus VALUES(5604,'フォッカチオ（モーニング）','Focaccia (Morning)','佛卡夏（早餐）',137,150,272,0.7,3,31,0,20,NULL);
INSERT INTO menus VALUES(5605,'焼きシナモンフォッカチオ','Grilled Cinnamon Focaccia','烤肉桂佛卡夏',182,200,339,0.8,3,31,0,20,NULL);
INSERT INTO menus VALUES(5606,'パンチェッタとチーズのパニーニ','Pancetta & Cheese Panini','培根奶酪帕尼尼',273,300,422,1.8,3,31,0,20,NULL);
INSERT INTO menus VALUES(5607,'フォッカチオ（モーニング単品）','Focaccia (Morning Single)','佛卡夏（早餐单品）',137,150,241,0.6,3,31,0,20,NULL);
INSERT INTO menus VALUES(5615,'焼きシナモンフォッカチオ コンビ','Grilled Cinnamon Focaccia Combo','烤肉桂佛卡夏套餐',273,300,339,0.8,3,31,0,20,NULL);
INSERT INTO menus VALUES(5616,'パンチェッタとチーズのパニーニ コンビ','Pancetta & Cheese Panini Combo','培根奶酪帕尼尼套餐',364,400,422,1.8,3,31,0,20,NULL);
INSERT INTO menus VALUES(5625,'焼きシナモンフォッカチオ セット','Grilled Cinnamon Focaccia Set','烤肉桂佛卡夏套餐',319,350,434,1.4,3,31,0,20,NULL);
INSERT INTO menus VALUES(5626,'パンチェッタとチーズのパニーニ セット','Pancetta & Cheese Panini Set','培根奶酪帕尼尼套餐',410,450,517,2.3,3,31,0,20,NULL);
INSERT INTO menus VALUES(5627,'フォッカチオ セット','Focaccia Set','佛卡夏套餐',273,300,367,1.3,3,31,0,20,NULL);
INSERT INTO menus VALUES(5630,'卵とパンチェッタとチーズのパニーニ','Egg, Pancetta & Cheese Panini','鸡蛋培根奶酪帕尼尼',319,350,422,1.8,3,31,0,20,NULL);
INSERT INTO menus VALUES(5631,'卵とパンチェッタとチーズのパニーニ コンビ','Egg, Pancetta & Cheese Panini Combo','鸡蛋培根奶酪帕尼尼套餐',410,450,422,1.8,3,31,0,20,NULL);
INSERT INTO menus VALUES(5632,'卵とパンチェッタとチーズのパニーニ セット','Egg, Pancetta & Cheese Panini Set','鸡蛋培根奶酪帕尼尼套餐',455,500,517,2.4,3,31,0,20,NULL);
INSERT INTO menus VALUES(7101,'ランチスープ','Lunch Soup','午餐汤',0,0,3,1.4,2,13,0,2,NULL);
INSERT INTO menus VALUES(7102,'ランチサラダ（トマトサラダ）','Lunch Salad (Tomato Salad)','午餐沙拉（番茄沙拉）',0,0,56,0.3,2,12,0,1,NULL);
INSERT INTO menus VALUES(7103,'ランチサラダ（コールスロー）','Lunch Salad (Coleslaw)','午餐沙拉（凉拌卷心菜）',0,0,129,0.7,2,12,0,1,NULL);
INSERT INTO menus VALUES(7104,'ランチサラダ（コーン）','Lunch Salad (Corn)','午餐沙拉（玉米）',0,0,61,0.4,2,12,0,1,NULL);
INSERT INTO menus VALUES(7105,'タラコソースシシリー風（ランチ）','Tarako Sauce Siciliano (Lunch)','鳕鱼籽酱西西里风味（午餐）',455,500,538,2.2,2,23,0,18,NULL);
INSERT INTO menus VALUES(7106,'パルマ風スパゲッティ（ランチ）','Spaghetti with Tomato & Pancetta (Lunch)','番茄培根意面（午餐）',455,500,726,3.6,2,23,0,18,NULL);
INSERT INTO menus VALUES(7107,'ミートソースボロニア風（ランチ）','Spaghetti Meat Sauce (Lunch)','意大利肉酱面（午餐）',455,500,584,3,2,23,0,18,NULL);
INSERT INTO menus VALUES(7108,'牛100％オニオンソースのハンバーグ','100% Beef Hamburg Steak with Onion Sauce','纯牛肉洋葱酱汉堡牛排',546,600,474,2.6,2,24,0,14,NULL);
INSERT INTO menus VALUES(7109,'牛100％ディアボラ風ハンバーグ','100% Beef Hamburg Steak Diavola','纯牛肉恶魔风汉堡牛排',546,600,595,3,2,24,0,14,NULL);
INSERT INTO menus VALUES(7110,'ライス（ランチ）','Rice (Lunch)','米饭（午餐）',0,0,303,0.1,2,31,0,19,NULL);
INSERT INTO menus VALUES(7111,'フォッカチオ（ランチ）','Focaccia (Lunch)','佛卡夏（午餐）',0,0,241,0.6,2,31,0,20,NULL);
INSERT INTO menus VALUES(7112,'ガーリックフォッカチオ（ランチ）','Garlic Focaccia (Lunch)','蒜香佛卡夏（午餐）',0,0,297,0.8,2,31,0,20,NULL);
INSERT INTO menus VALUES(8101,'辛味チキン（テイクアウト）','Spicy Grilled Chicken (Takeout)','香辣鸡翅（外带）',273,300,295,1.8,5,14,0,3,NULL);
INSERT INTO menus VALUES(8102,'バッファローモッツァレラのマルゲリータピザ（テイクアウト）','Buffalo Mozzarella Margherita Pizza (Takeout)','水牛马苏里拉披萨（外带）',364,400,539,2.4,5,22,0,17,NULL);
INSERT INTO menus VALUES(8103,'たっぷりコーンのピザ（テイクアウト）','Corn Pizza (Takeout)','足量甜玉米比萨（外带）',364,400,568,2.8,5,22,0,17,NULL);
INSERT INTO menus VALUES(8104,'ソーセージピザ（テイクアウト）','Sausage Pizza (Takeout)','香肠比萨（外带）',364,400,643,3.9,5,22,0,17,NULL);
INSERT INTO menus VALUES(8105,'野菜ときのこのピザ（テイクアウト）','Vegetable & Mushroom Pizza (Takeout)','蔬菜蘑菇比萨（外带）',364,400,580,2.9,5,22,0,17,NULL);
INSERT INTO menus VALUES(8106,'ミラノ風ドリア（テイクアウト）','Meat Sauce Doria (Takeout)','米兰风肉酱多利亚饭（外带）',273,300,550,2.7,5,21,0,7,NULL);
INSERT INTO menus VALUES(8107,'冷凍イタリアンプリン（1個）','Frozen Italian Cream Caramel (1pc)','冷冻意大利布丁（1个）',218,240,223,0.2,5,32,0,23,NULL);
INSERT INTO menus VALUES(8108,'冷凍辛味チキン（1.5kg）','Frozen Spicy Grilled Chicken (1.5kg)','冷冻香辣鸡翅（1.5kg）',2000,2200,3690,22.1,5,14,0,3,NULL);
INSERT INTO menus VALUES(8109,'サイゼリヤドレッシング（500ml）','SAIZERIYA Dressing (500ml)','SAIZERIYA沙拉酱（500ml）',455,500,2430,11.9,5,53,0,31,NULL);
INSERT INTO menus VALUES(8110,'エクストラ・バージン・オリーブオイル（テイクアウト）','Extra Virgin Olive Oil (Takeout)','特级初榨橄榄油（外带）',1091,1200,4050,0,5,53,0,31,NULL);
INSERT INTO menus VALUES(9101,'あさりとトマトのアクアパッツァ','Clams and Tomato Acqua Pazza','蛤蜊番茄意式水煮鱼',409,450,257,2.2,6,14,0,7,NULL);
INSERT INTO menus VALUES(9102,'チェリーソースのミルクジェラート','Milk Gelato with Classic Cherry Sauce','樱桃酱配牛奶冰激淋',409,450,179,0.1,6,32,0,22,NULL);
INSERT INTO menus VALUES(9103,'チェリーソースのティラミス','Tiramisu with Classic Cherry Sauce','樱桃酱配提拉米苏',455,500,286,0.2,6,32,0,28,NULL);
INSERT INTO menus VALUES(9104,'白桃ソースのミルクジェラート','Milk Gelato with White Peach Sauce','白桃酱配牛奶冰激淋',409,450,167,0.1,6,32,0,22,NULL);
INSERT INTO menus VALUES(9105,'白桃ソースのティラミス','Tiramisu with Classic White Peach Sauce','白桃酱配提拉米苏',455,500,273,0.2,6,32,0,28,NULL);
INSERT INTO menus VALUES(9106,'プロフィットロール','Profiteroles','泡芙',318,350,226,0.3,6,32,0,24,NULL);
INSERT INTO menus VALUES(9107,'プロフィットロール＆ミルクジェラート','Profiteroles & Milk Gelato','泡芙配牛奶冰激淋',500,550,346,0.4,6,32,0,22,NULL);
INSERT INTO menus VALUES(9108,'アップルタルト','Apple Tart','苹果塔',318,350,144,0.2,6,32,0,24,NULL);
INSERT INTO menus VALUES(9109,'アップルタルト＆ミルクジェラート','Apple Tart & Milk Gelato','苹果塔配牛奶冰激淋',500,550,265,0.2,6,32,0,22,NULL);
PRAGMA writable_schema=ON;
CREATE TABLE IF NOT EXISTS sqlite_sequence(name,seq);
DELETE FROM sqlite_sequence;
INSERT INTO sqlite_sequence VALUES('categories',6);
INSERT INTO sqlite_sequence VALUES('icons',31);
INSERT INTO sqlite_sequence VALUES('categories',6);
INSERT INTO sqlite_sequence VALUES('icons',31);
PRAGMA writable_schema=OFF;
COMMIT;
