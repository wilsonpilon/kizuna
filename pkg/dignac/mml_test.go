package dignac

import "testing"

func TestParseMMLBasicScale(t *testing.T) {
	events, err := parseMML("O4 C D E F G A B >C")
	if err != nil {
		t.Fatalf("parseMML falhou: %v", err)
	}
	if len(events) != 8 {
		t.Fatalf("esperava 8 eventos, obteve %d", len(events))
	}
	// C4 = (4-2)*12 + 0 = 24
	if events[0].NoteIdx != 24 {
		t.Errorf("C4: esperava índice 24, obteve %d", events[0].NoteIdx)
	}
	// D4 = 24 + 2 = 26
	if events[1].NoteIdx != 26 {
		t.Errorf("D4: esperava índice 26, obteve %d", events[1].NoteIdx)
	}
	// B4 = 24 + 11 = 35
	if events[6].NoteIdx != 35 {
		t.Errorf("B4: esperava índice 35, obteve %d", events[6].NoteIdx)
	}
	// >C -- sobe uma oitava antes da nota: C5 = (5-2)*12 = 36
	if events[7].NoteIdx != 36 {
		t.Errorf(">C: esperava índice 36 (C5), obteve %d", events[7].NoteIdx)
	}
}

func TestParseMMLSharpFlat(t *testing.T) {
	events, err := parseMML("O4 C+ D-")
	if err != nil {
		t.Fatalf("parseMML falhou: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("esperava 2 eventos, obteve %d", len(events))
	}
	// C#4 = 24 + 1 = 25
	if events[0].NoteIdx != 25 {
		t.Errorf("C+4: esperava índice 25, obteve %d", events[0].NoteIdx)
	}
	// D-4 = 26 - 1 = 25 (mesmo semitom que C#4)
	if events[1].NoteIdx != 25 {
		t.Errorf("D-4: esperava índice 25, obteve %d", events[1].NoteIdx)
	}
}

func TestParseMMLDurationAndDot(t *testing.T) {
	events, err := parseMML("L4 C C8 C4. R4")
	if err != nil {
		t.Fatalf("parseMML falhou: %v", err)
	}
	if len(events) != 4 {
		t.Fatalf("esperava 4 eventos, obteve %d", len(events))
	}
	// L4 -> duração base 50
	if events[0].Duration != 50 {
		t.Errorf("C (L4): esperava duração 50, obteve %d", events[0].Duration)
	}
	// C8 (colcheia) -> metade de L4 = 25
	if events[1].Duration != 25 {
		t.Errorf("C8: esperava duração 25, obteve %d", events[1].Duration)
	}
	// C4. (pontuada) -> 50 * 1.5 = 75
	if events[2].Duration != 75 {
		t.Errorf("C4.: esperava duração 75, obteve %d", events[2].Duration)
	}
	// R4 -> pausa, volume 0, duração 50
	if events[3].Volume != 0 || events[3].Duration != 50 {
		t.Errorf("R4: esperava volume 0 duração 50, obteve volume %d duração %d", events[3].Volume, events[3].Duration)
	}
}

func TestParseMMLVolume(t *testing.T) {
	events, err := parseMML("V15 C V0 D")
	if err != nil {
		t.Fatalf("parseMML falhou: %v", err)
	}
	if events[0].Volume != 15 {
		t.Errorf("esperava volume 15, obteve %d", events[0].Volume)
	}
	if events[1].Volume != 0 {
		t.Errorf("esperava volume 0, obteve %d", events[1].Volume)
	}
}

func TestParseMMLOctaveOutOfRange(t *testing.T) {
	if _, err := parseMML("O1 C"); err == nil {
		t.Fatal("esperava erro para oitava fora do alcance (O1)")
	}
	if _, err := parseMML("O7 C"); err == nil {
		t.Fatal("esperava erro para oitava fora do alcance (O7)")
	}
	if _, err := parseMML("O2 <C"); err == nil {
		t.Fatal("esperava erro ao descer abaixo da oitava mínima com '<'")
	}
}

func TestParseMMLInvalidChar(t *testing.T) {
	if _, err := parseMML("O4 CZ"); err == nil {
		t.Fatal("esperava erro para caractere MML inválido 'Z'")
	}
}

func TestEncodeEvents(t *testing.T) {
	events := []psgEvent{
		{Channel: 0, NoteIdx: 24, Volume: 12, Duration: 50},
	}
	out := encodeEvents(events)
	want := "    DB 00h, 18h, 0Ch, 32h\n    DB 0FFh\n"
	if out != want {
		t.Errorf("encodeEvents = %q, esperado %q", out, want)
	}
}
