package game

type CardType int

const (
	PASS_CARDS     CardType = -2 //过
	NO_CARDS       CardType = -1 //前面还没有牌（首家）
	ERROR_CARDS    CardType = 0  //错误牌型
	SINGLE_CARD    CardType = 1  //单牌
	DOUBLE_CARD    CardType = 2  //对子
	THREE_CARD     CardType = 3  //3不带
	THREE_ONE_CARD CardType = 4  //3带1
	THREE_TWO_CARD CardType = 5  //3带2
	BOMB_TWO_CARD  CardType = 6  //4带2
	STRAIGHT       CardType = 7  //连牌
	CONNECT_CARD   CardType = 8  //连对
	AIRCRAFT       CardType = 9  //飞机不带
	AIRCRAFT_CARD  CardType = 10 //飞机带单牌
	AIRCRAFT_WING  CardType = 11 //飞机带对子
	BOMB_CARD      CardType = 12 //炸弹
	KINGBOMB_CARD  CardType = 13 //王炸
)
