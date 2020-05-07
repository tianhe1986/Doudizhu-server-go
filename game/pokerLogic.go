package game

import "sort"

// 其实现在是空的，但是万一有用呢？所以就没有写成静态方法了
type PokerLogic struct {

}

// 计算牌型
func (p *PokerLogic) CalcuPokerType(cards []int) CardType {
	// 转换为点数
	points := p.CardsToPoints(cards)

	length := len(points)

	if length == 1 { // 一张牌，当然是单牌
		return SINGLE_CARD
	} else if length == 2 { // 两张牌，王炸或一对
		if points[0] == 16 && points[1] == 17 { // 王炸
			return KINGBOMB_CARD
		}
		if points[0] == points[1] { // 对子
			return DOUBLE_CARD
		}
	} else if length == 3 { // 三张，只检查三不带
		if points[0] == points[1] && points[1] == points[2] {
			return THREE_CARD
		}
	} else if length == 4 { // 四张，炸弹或三带一
		maxSameNum := p.CalcuMaxSameNum(points)
		if maxSameNum == 4 { // 四张相同，炸弹
			return BOMB_CARD
		}

		if maxSameNum == 3 { // 三张点数相同的，3带1
			return THREE_ONE_CARD
		}
	} else if length >= 5 && p.IsStraight(points) && points[length - 1] < 15 { // 大于等于5张，是点数连续，且最大点数不超过2， 则是顺子
		return STRAIGHT
	} else if length == 5 { // 5张，只需检查3带2
		// 最多有3张相等的，且只有两种点数，则是3带2
		if p.CalcuMaxSameNum(points) == 3 && p.CalcuDiffPoint(points) == 2 {
			return THREE_TWO_CARD
		}
	} else { // 大于6的情况，再分别判断
		maxSameNum := p.CalcuMaxSameNum(points)
		diffPointNum := p.CalcuDiffPoint(points)

		// length能被3整除， 最大相同数量是3， 不同点数是length/3, 且最大与最小点数相差 length/3 - 1， 则是连续三张
		if length % 3 == 0 && maxSameNum == 3 && diffPointNum == length / 3 && (points[length - 1] - points[0] == length/3 - 1) && points[length - 1] < 15 {
			return AIRCRAFT
		}

		// 与上面连续三张判断类似，连对
		if length % 2 == 0 && maxSameNum == 2 && diffPointNum == length / 2 && (points[length - 1] - points[0] == length/2 - 1) && points[length - 1] < 15 {
			return CONNECT_CARD
		}

		// 飞机三带一
		if length % 4 == 0 {
			// 获得所有数量为3的点数
			threePoints := p.GetSameNumPoints(points, 3)
			if len(threePoints) == length/4 && p.IsStraight(threePoints) && threePoints[len(threePoints) - 1] < 15 {
				return AIRCRAFT_CARD
			}
		}

		// 飞机三带二
		if length % 5 == 0 {
			threePoints := p.GetSameNumPoints(points, 3)
			// 三带二里面，不会出现单牌的情况
			onePoints := p.GetSameNumPoints(points, 1)
			if len(onePoints) == 0 && len(threePoints) == length/5 && p.IsStraight(threePoints) && threePoints[len(threePoints) - 1] < 15 {
				return AIRCRAFT_WING
			}
		}

		// 四带二
		if length == 6 {
			if maxSameNum == 4 {
				return BOMB_TWO_CARD
			}
		}

		// 四带两对
		if length == 8 {
			if maxSameNum == 4 {
				twoPoints := p.GetSameNumPoints(points, 2)
				if len(twoPoints) == 2 {
					return BOMB_FOUR_CARD
				}
			}
		}
		// 连续四带二
		if length % 6 == 0 {
			fourPoints := p.GetSameNumPoints(points, 4)
			if len(fourPoints) == length/6 && p.IsStraight(fourPoints) && fourPoints[len(fourPoints) - 1] < 15 {
				return BOMB_TWO_STRAIGHT_CARD
			}
		}
		// 连续四带两对
		if length % 8 == 0 {
			fourPoints := p.GetSameNumPoints(points, 4)
			twoPoints := p.GetSameNumPoints(points, 2)
			if len(twoPoints) == 2 * len(fourPoints) && len(fourPoints) == length/8 && p.IsStraight(fourPoints) && fourPoints[len(fourPoints) - 1] < 15 {
				return BOMB_FOUR_STRAIGHT_CARD
			}
		}
	}

	// 没有这个牌型的
	return ERROR_CARDS
}

// 取出所有点数数量等于num的点数
// 例如，现在牌中有3个3，3个4，2个5，1个6， 取出数量等于3的点数，则返回[3, 4]，取出数量等于2的点数，则返回[5]，取出数量等于1的点数，则返回[6]，其他都返回空数组
func (p *PokerLogic) GetSameNumPoints(points []int, num int) []int {
	length := len(points)
	newPoints := make([]int, length)
	pointIndex := 0

	nowNum := 1

	for i := 1; i < length; i++ {
		if points[i] == points[i-1] { // 与前一张相同
			nowNum++
		} else { // 与前一张不同，若前一张出现num次，加入数组
			if nowNum == num {
				newPoints[pointIndex] = points[i-1]
				pointIndex++
			}
			nowNum = 1
		}
	}

	if nowNum == num {
		newPoints[pointIndex] = points[length-1]
		pointIndex++
	}

	return newPoints[0:pointIndex]
}

// 是否是顺子
func (p *PokerLogic) IsStraight(points []int) bool {
	length := len(points)
	for i := 1; i < length; i++ {
		if points[i] != points[i-1] + 1 { // 与前一张相同
			return false
		}
	}

	return true
}

// 有多少种不同的点数
func (p *PokerLogic) CalcuDiffPoint(points []int) int {
	diffNum := 1

	length := len(points)
	for i := 1; i < length; i++ {
		if points[i] != points[i-1] { // 与前一张不同，则出现了新的点数
			diffNum++
		}
	}

	return diffNum
}

// 最多有几张点数相等的牌
func (p *PokerLogic) CalcuMaxSameNum(points []int) int {
	length := len(points)
	nowNum := 1
	maxNum := 1

	for i := 1; i < length; i++ {
		if points[i] == points[i-1] { // 与前一张相同
			nowNum++
		} else { // 与前一张不同，重新开始计数
			if nowNum > maxNum {
				maxNum = nowNum
			}
			nowNum = 1
		}
	}

	if nowNum > maxNum {
		maxNum = nowNum
	}

	return maxNum
}

// 牌id转点数
func (p *PokerLogic) CardsToPoints(cards []int) []int {
	length := len(cards)
	points := make([]int, length)

	var point int
	for i := 0; i < length; i++ {
		value := cards[i]
		if value < 53 { // id 1-4 对应的是3， 5-8对应4， 依此类推 45 - 48对应14(A)， 49-52对应15(2)
			if value % 4 == 0 {
				point = value / 4 + 2
			} else {
				point = value / 4 + 3
			}
		} else { // 小王和大王
			point = value /4 + 2 + value % 4
		}

		points[i] = point
	}

	// 按点数升序排序
	sort.Ints(points)

	return points
}

// 计算头牌
func (p *PokerLogic) CalcuPokerHeader(cards []int, cardType CardType) int {
	points := p.CardsToPoints(cards)

	switch cardType {
	case SINGLE_CARD, DOUBLE_CARD, THREE_CARD, STRAIGHT, CONNECT_CARD, AIRCRAFT, BOMB_CARD:
		return points[0]
	case THREE_ONE_CARD,THREE_TWO_CARD,BOMB_TWO_CARD:
		return points[2]
	case AIRCRAFT_CARD,AIRCRAFT_WING:
		return p.FirstPoint(points, 3)
	}

	return 0
}

// 获得首个出现num次的点数
func (p *PokerLogic) FirstPoint(points []int, num int) int {
	nowNum := 1;
	length := len(points)

	for i := 1; i < length; i++ {
		if (points[i] == points[i-1]) { //与上一张相同，数量加1
			nowNum++;
		} else { //重新开始计算
			if (nowNum == num) {
				return points[i-1];
			}
			nowNum = 1;
		}
	}

	if (nowNum == num) {
		return points[length - 1];
	}

	return 0
}

// 是否可以出牌
func (p *PokerLogic) CanOut(newCardSet *CardSet, nowCardSet *CardSet) bool {
	// 当前是第一次出牌，牌型正确即可
	if nowCardSet.Type == NO_CARDS && newCardSet.Type != ERROR_CARDS {
		return true
	}

	// 王炸，天下第一
	if newCardSet.Type == KINGBOMB_CARD {
		return true
	}

	// 炸弹，检查前面是不是也是炸弹
	if newCardSet.Type == BOMB_CARD {
		if nowCardSet.Type == BOMB_CARD {
			return newCardSet.Header > nowCardSet.Header
		} else {
			return true
		}
	}

	// 同类型，张数相同，头牌更大
	if newCardSet.Type == nowCardSet.Type && len(newCardSet.Cards) == len(nowCardSet.Cards) && newCardSet.Header > nowCardSet.Header {
		return true
	}

	return false
}
