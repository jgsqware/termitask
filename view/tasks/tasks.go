package tasks

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
	"github.com/jgsqware/termitask/view"
	tui "github.com/marcusolsson/tui-go"
)

type task struct {
	Donetime time.Time
	Text     string
}

type tasks []task

func (t *tasks) encode() ([]byte, error) {
	enc, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decode(data []byte) (*tasks, error) {
	var t *tasks
	err := json.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func initBucket(db *bolt.DB, bucketName string) {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

func getTasks(db *bolt.DB, bucketName string) *tasks {
	t := &tasks{}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get([]byte("tasks"))
		var err error
		t, err = decode(v)
		return err
	})
	return t
}

func saveTasks(db *bolt.DB, bucketName string, tasks tasks) {
	err := db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(bucketName))
		t, err := tasks.encode()
		if err != nil {
			log.Fatal(err)
		}
		err = b.Put([]byte("tasks"), t)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})

	if err != nil {

	}
}

func (t task) String() string {
	if t.Donetime.IsZero() {
		return "[ ] " + t.Text
	}
	return "[✓] " + t.Text + " - done " + t.Donetime.Format("02/01/2006 15:04")
}

func onItemActivated(db *bolt.DB, bucketName string) func(l *tui.List) {
	return func(l *tui.List) {
		s := l.Selected()
		if s >= 0 {
			ts := *getTasks(db, bucketName)
			if ts[s].Donetime.IsZero() {
				ts[s].Donetime = time.Now()
			} else {
				ts[s].Donetime = time.Time{}
			}
			draw(ts, l)
			l.Select(s)
			saveTasks(db, bucketName, ts)
		}
	}
}

func onSubmit(db *bolt.DB, bucketName string, l *tui.List) func(e *tui.Entry) {
	return func(e *tui.Entry) {
		if e.Text() != "" {
			t := task{Text: e.Text()}
			ts := *getTasks(db, bucketName)
			ts = append(ts, t)
			draw(ts, l)
			e.SetText("")
			saveTasks(db, bucketName, ts)
		}
	}
}

func initTaskList(db *bolt.DB, bucketName string, l *tui.List) {
	ts := getTasks(db, bucketName)

	if ts != nil {
		for _, t := range *ts {
			l.AddItems(t.String())
		}
	} else {
		saveTasks(db, bucketName, tasks{})
	}
}

func legendShortcut(kb string) tui.Widget {
	k := tui.NewLabel("<" + kb + ">")
	k.SetStyleName(view.StyleShortcut)
	return k
}

func NewTaskBox(ui view.UI, name string, kbList string, kbEdit string, kbClear string) *tui.Box {
	initBucket(ui.Db, name)
	l := tui.NewList()
	initTaskList(ui.Db, name, l)
	l.OnItemActivated(onItemActivated(ui.Db, name))
	ui.AddWidget(l, kbList)

	legend := tui.NewHBox(
		legendShortcut(kbList),
		tui.NewPadder(1, 0, tui.NewLabel("Show list")),
		legendShortcut(kbEdit),
		tui.NewPadder(1, 0, tui.NewLabel("Add task")),
		legendShortcut(kbClear),
		tui.NewPadder(1, 0, tui.NewLabel("Delete done task")),
		tui.NewSpacer(),
	)

	lb := tui.NewVBox(l)
	lb.SetTitle(name)
	lb.SetBorder(true)
	lb.SetSizePolicy(tui.Expanding, tui.Expanding)

	e := tui.NewEntry()
	e.SetSizePolicy(tui.Expanding, tui.Maximum)
	e.OnSubmit(onSubmit(ui.Db, name, l))
	ui.AddWidget(e, kbEdit)

	tb := tui.NewHBox(e)
	tb.SetTitle("+ " + name)
	tb.SetBorder(true)
	tb.SetSizePolicy(tui.Expanding, tui.Maximum)

	ui.SetKeybinding(kbClear, func() {
		clearDone(ui.Db, name, l)
	})
	return tui.NewVBox(lb, tb, legend)
}

func clearDone(db *bolt.DB, bucketName string, l *tui.List) {
	tl := *getTasks(db, bucketName)
	vtl := make(tasks, 0)
	for _, t := range tl {
		if t.Donetime.IsZero() {
			vtl = append(vtl, t)
		}
	}
	saveTasks(db, bucketName, vtl)
	draw(vtl, l)
}

func draw(tl tasks, l *tui.List) {
	l.RemoveItems()
	for _, t := range tl {
		l.AddItems(t.String())
	}
}
