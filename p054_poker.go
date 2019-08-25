package main

import ("fmt"
        "sort"
        //"reflect"     // debug data type :: reflect.TypeOf(<object>)
        "bufio"
        "log"
        "os"
        "strings"
        "time"
)

// 2 character notation ... 1st the value, then the suit
type Card struct {
  val  int
  suit string
}

// best combination in hand
type Hand struct {
  points int      // assign each winning combo a value
  winner string   // the winning combo itself
  win_value int   // the (highest) card value of the winning combo
}

// data structure for analyzing the cards
type CardAnalysis struct {
  same_suit bool
  in_sequence bool
  high_card int
  card_count map[int]int    // histogram of the cards
}

// leveraging Go's built in sort mechanism
type ByPoker []Card
func (a ByPoker) Len() int { return len(a) }
func (a ByPoker) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPoker) Less(i, j int) bool {
  switch {
    case a[i].val != a[j].val:
      return a[i].val < a[j].val
    default:
      return a[i].suit < a[j].suit
  }
}

// read and map the 5 card hand
func GetHand(h []string) []Card {
  values := map[string]int {
    "2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7, "8": 8, "9": 9,
    "T": 10, "J": 11, "Q": 12, "K": 13, "A": 14,
  }
  suits := map[string]string {
    "C": "Clubs", "D": "Diamonds", "H": "Hearts", "S": "Spades",
  }

  var p []Card
  for _, c := range h {
    card := Card {val: values[c[:1]], suit: suits[c[1:2]]}
    p = append(p, card)
    //fmt.Printf("%s :: %v\n", c, card)
  }
  return p
}

// simple search routine to see if card was seen before
func HaveSeenCard(k int, stack map[int]int) bool {
  _, ok := stack[k];
  return ok
}

// analyze cards and save details to determine winning combo
func SaveDetails(hand []Card) CardAnalysis{
  prev_val := hand[0].val - 1
  var details CardAnalysis
  details.same_suit = true ; details.in_sequence = true ; details.high_card = 0
  details.card_count = make(map[int]int)
  for _, card := range hand {
    // cards of the same suit?
    if details.same_suit && card.val != hand[0].val {
      details.same_suit = false
    }
    // cards in sequence?
    if details.in_sequence && card.val == prev_val+1 {
      prev_val = card.val
    } else {
      details.in_sequence = false
    }
    // highest value
    if card.val > details.high_card {
      details.high_card = card.val
    }
    // count/histogram of each value
    if HaveSeenCard(card.val, details.card_count) {
      details.card_count[card.val] += 1
      }  else {
        details.card_count[card.val] = 1
      }
  }
  return details
}

// determine the best combination ...
// early return means only the best combination is analyzed.
// if there is a tie... we should ideally determine the next best combo
func WinningCombo(details CardAnalysis) Hand {
  my_len := len(details.card_count)
  switch {
    // Royal FLush, Straight Flush or Flush
    case details.same_suit:
      switch {
        case details.in_sequence:
          if details.high_card == 14 {
            return Hand{10, "Royal Flush", 14}
          } else {
            return Hand{9, "Straight Flush", details.high_card}
          }
        default:
          return Hand{6, "Flush", details.high_card}
      }
    // Straight .... in sequence but mixed suit
    case details.in_sequence:
      return Hand{5, "Straight", details.high_card}
    default:
      switch {
        // Four of a Kind or Full House
        case my_len == 2:
          for k, v := range details.card_count {
            if v == 4 {
              return Hand{8, "Four of a Kind", k}
            }
            if v == 3 {
              return Hand{7, "Full House", k}
            }
          }
        // Three of a Kind or Two Pairs
        case my_len == 3:
          for k, v := range details.card_count {
            if v == 3 {
              return Hand{4, "Three of a Kind", k}
            }
            if v == 2 {
              return Hand{3, "Two Pairs", k}
            }
          }
        // The remaining two -- One Pair and High Card
        case my_len == 4:
          for k, v := range details.card_count {
            if v == 2 {
              return Hand{2, "One Pair", k}
            }
          }
        default:
          return Hand{1, "High Card", details.high_card}
      }
  }
  return Hand{0, "Dummy", 0}
}

// Decode hand, Sort the cards, Analyze it, and return the best combination
func AnalyzeHand(h []string) Hand {
  p := GetHand(h)
  sort.Sort(ByPoker(p))
  return(WinningCombo(SaveDetails(p)))
}

// Declare the winner :: Logic not complete if there is a deadlock
func DecideWinner(h1, h2 Hand) int{
  //fmt.Println("player 1 : ", h1)
  //fmt.Println("player 2 : ", h2)
  switch {
  case h1.points > h2.points:
    //fmt.Println("Player 1 wins : ", h1.winner, "  BEATS  ", h2.winner)
    return 1
  case h1.points < h2.points:
    //fmt.Println("Player 2 wins : ", h1.winner, "BEATEN BY", h2.winner)
    return 2
  default:
    switch {
    case h1.win_value > h2.win_value:
      //fmt.Println("Player 1 wins : ", h1.win_value, "  BEATS  ", h2.win_value)
      return 1
    case h1.win_value < h2.win_value:
      //fmt.Println("Player 2 wins : ", h1.win_value, "BEATEN BY", h2.win_value)
      return 2
    default:
      // need to continue analysis to determine next winning combo
      fmt.Println("More analysis Needed : ", h1, h2)
      return 0
    }
  }
}

func ReadGamesFromFile(f string) map[int][]string{
  file, err := os.Open(f)
  if err != nil {log.Fatal(err)}
  defer file.Close()

  games := make(map[int][]string)
  cnt := 0
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    game := scanner.Text()
    games[cnt] = strings.Fields(game)
    cnt++
  }

  if err := scanner.Err(); err != nil {
    log.Fatal(err)
  }
  return games
}


func main() {
  t1 := time.Now()
  var p1_win, p2_win int

  games := ReadGamesFromFile("./p54_poker.txt")
  for idx, game := range games {
    //game := []string {"4C", "2S", "AD", "TH", "9D", "8S", "3S", "KD", "3C", "5H"}
    //game := []string {"4C", "2S", "AD", "TH", "9D", "3H", "3S", "KD", "3C", "3D"}

    player1 := game[0:5]
    player2 := game[5:]
    h1 := AnalyzeHand(player1)
    h2 := AnalyzeHand(player2)
    winner := DecideWinner(h1, h2)

    switch {
    case winner == 1:
      p1_win++
    case winner == 2:
      p2_win++
    default:
      fmt.Println("Manually Decide :: Game =", idx, " :: ", player1, " vs ", player2)
    }
  }
  fmt.Printf("\nFINAL :: P1 win = %d   ::  P2 win = %d  ::  Manual = %d\n",
    p1_win, p2_win, len(games)-p1_win-p2_win)

  fmt.Println("Time taken : ", time.Since(t1))
}
