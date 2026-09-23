package dignac

import (
	"fmt"
	"strings"
)

// psgEvent representa um evento de PSG_PlaySequence: canal, índice de nota
// (0..59 em PSG_NoteTable), volume (0..15) e duração (contador de laço,
// não milissegundos -- mesma convenção de lib/src/psg/*.asm).
type psgEvent struct {
	Channel  int
	NoteIdx  int
	Volume   int
	Duration int
}

// noteOffsets mapeia a letra da nota (A..G) para o deslocamento semitônico
// dentro da oitava, na mesma ordem usada por PSG_NoteTable (C=0 .. B=11).
var noteOffsets = map[byte]int{
	'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11,
}

const (
	mmlMinOctave = 2
	mmlMaxOctave = 6
	// duração de referência (em unidades de PSG_PlaySequence) para uma
	// semínima (L4) -- mesmo valor usado nos exemplos manuais de
	// sample/music/main.asm, mantém o "tempo" comparável.
	mmlQuarterDuration = 50
)

// parseMML traduz uma string MML (Music Macro Language, o mesmo dialeto que
// o PLAY do MSX-BASIC usa) em uma sequência de eventos PSG_PlaySequence.
// Subconjunto suportado: notas A-G com # /+ (sustenido) ou - (bemol) e
// duração numérica opcional (com '.' de aumento); O<n>/</> (oitava);
// L<n> (duração padrão); V<n> (volume padrão); R[n] (pausa). Erro claro para
// qualquer caractere ou combinação não reconhecida -- em vez de silenciar,
// como um Assembler faria (ver o histórico de bugs deste projeto).
func parseMML(src string) ([]psgEvent, error) {
	octave := 4
	length := 4 // L4 = semínima, como no MSX-BASIC
	volume := 12
	var events []psgEvent

	i := 0
	n := len(src)
	for i < n {
		ch := src[i]
		switch {
		case ch == ' ' || ch == '\t':
			i++

		case ch == '<':
			octave--
			if octave < mmlMinOctave {
				return nil, fmt.Errorf("MML: oitava abaixo de O%d (posição %d)", mmlMinOctave, i)
			}
			i++

		case ch == '>':
			octave++
			if octave > mmlMaxOctave {
				return nil, fmt.Errorf("MML: oitava acima de O%d (posição %d)", mmlMaxOctave, i)
			}
			i++

		case ch == 'O' || ch == 'o':
			val, next, err := mmlReadNumber(src, i+1)
			if err != nil || val < mmlMinOctave || val > mmlMaxOctave {
				return nil, fmt.Errorf("MML: 'O' precisa de uma oitava entre %d e %d (posição %d)", mmlMinOctave, mmlMaxOctave, i)
			}
			octave = val
			i = next

		case ch == 'L' || ch == 'l':
			val, next, err := mmlReadNumber(src, i+1)
			if err != nil || val <= 0 {
				return nil, fmt.Errorf("MML: 'L' precisa de uma duração positiva (posição %d)", i)
			}
			length = val
			i = next

		case ch == 'V' || ch == 'v':
			val, next, err := mmlReadNumber(src, i+1)
			if err != nil || val < 0 || val > 15 {
				return nil, fmt.Errorf("MML: 'V' precisa de um volume entre 0 e 15 (posição %d)", i)
			}
			volume = val
			i = next

		case ch == 'R' || ch == 'r':
			dur, next, err := mmlReadDuration(src, i+1, length)
			if err != nil {
				return nil, err
			}
			events = append(events, psgEvent{Channel: 0, NoteIdx: 0, Volume: 0, Duration: dur})
			i = next

		case isNoteLetter(ch):
			offset, ok := noteOffsets[toUpperByte(ch)]
			if !ok {
				return nil, fmt.Errorf("MML: nota inválida '%c' (posição %d)", ch, i)
			}
			i++
			if i < n && (src[i] == '+' || src[i] == '#') {
				offset++
				i++
			} else if i < n && src[i] == '-' {
				offset--
				i++
			}
			if offset < 0 || offset > 11 {
				return nil, fmt.Errorf("MML: acidente leva a nota '%c' para fora da oitava (posição %d)", ch, i)
			}
			dur, next, err := mmlReadDuration(src, i, length)
			if err != nil {
				return nil, err
			}
			noteIdx := (octave-mmlMinOctave)*12 + offset
			events = append(events, psgEvent{Channel: 0, NoteIdx: noteIdx, Volume: volume, Duration: dur})
			i = next

		default:
			return nil, fmt.Errorf("MML: caractere não reconhecido '%c' (posição %d)", ch, i)
		}
	}

	return events, nil
}

// mmlReadDuration lê o sufixo numérico opcional (duração explícita) e o '.'
// opcional (aumenta 50%) após uma nota ou um R, convertendo a duração em
// unidades de comprimento de nota MML (4=semínima, 8=colcheia, ...) para
// unidades de PSG_PlaySequence.
func mmlReadDuration(src string, pos int, defaultLength int) (int, int, error) {
	noteLen := defaultLength
	next := pos
	if val, after, err := mmlReadNumber(src, pos); err == nil {
		noteLen = val
		next = after
	}
	if noteLen <= 0 {
		return 0, 0, fmt.Errorf("MML: duração de nota precisa ser positiva (posição %d)", pos)
	}

	dotted := false
	if next < len(src) && src[next] == '.' {
		dotted = true
		next++
	}

	// duração em unidades de PSG_PlaySequence: proporcional a 1/noteLen de
	// uma semibreve (L1), calibrada para que L4 (semínima) valha
	// mmlQuarterDuration -- mesma escala usada nos exemplos manuais.
	dur := (mmlQuarterDuration * 4) / noteLen
	if dotted {
		dur = dur + dur/2
	}
	if dur < 1 {
		dur = 1
	}
	if dur > 255 {
		dur = 255
	}
	return dur, next, nil
}

// mmlReadNumber lê um inteiro decimal simples a partir de pos, devolvendo o
// valor e o índice logo após o último dígito.
func mmlReadNumber(src string, pos int) (int, int, error) {
	start := pos
	for pos < len(src) && src[pos] >= '0' && src[pos] <= '9' {
		pos++
	}
	if pos == start {
		return 0, pos, fmt.Errorf("MML: número esperado na posição %d", start)
	}
	val := 0
	for _, c := range src[start:pos] {
		val = val*10 + int(c-'0')
	}
	return val, pos, nil
}

func isNoteLetter(ch byte) bool {
	u := toUpperByte(ch)
	return u >= 'A' && u <= 'G'
}

func toUpperByte(ch byte) byte {
	if ch >= 'a' && ch <= 'z' {
		return ch - 32
	}
	return ch
}

// encodeEvents converte os eventos em uma diretiva DB pronta para o corpo do
// Assembly gerado (4 bytes por evento + 0FFh de terminador).
func encodeEvents(events []psgEvent) string {
	var sb strings.Builder
	for _, e := range events {
		sb.WriteString(fmt.Sprintf("    DB %02Xh, %02Xh, %02Xh, %02Xh\n", e.Channel, e.NoteIdx, e.Volume, e.Duration))
	}
	sb.WriteString("    DB 0FFh\n")
	return sb.String()
}
