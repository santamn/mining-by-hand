# Difficultyについて

## Difficulty, Target, Bits

- Difficulty: マイニングの難易度。一つのブロックの生成が大体10分で完了するように調整される。
- Target: 許容されるブロックヘッダのハッシュ値の上限。Difficultyによって決まる。
- Bits: Targetの圧縮表現

## 計算方法

以下のように定数を定める。

- $T_{max}=$ `0xffff << 208`
- $D_0 = 1$(多分)

この時、 $D_{now}$ が現在のDifficulty、 $D_{new}$ が次のDifficulty、 $T$ が現在のTargetとすると、次のように計算が行われる。

$D_{new} = D_{now} \times \{2016 \times 10 / 実際に2016ブロックのマイニングに掛かった時間[分]\}$

$T = T_{max} / D_{now}$

また、Bitsは4byteの値であり、上位1byteが指数部(= index)、下位3byteが仮数部(= coefficient)と呼ばれる。この時、BitsとTargetの関係は以下のようになる。

$T = coefficient \times 2^{8 \times (index - 3)}$