package raft

import (
	"6.824/labgob"
	"bytes"
	"log"
)

//
// 将 Raft 的持久状态保存到稳定存储中，
// 以后可以在崩溃和重新启动后检索它。
// 参见论文的图 2 了解什么应该是持久化的。
//
func (rf *Raft) stateEncode() []byte {

	if rf == nil {
		// log.Printf("S[%d] Eh? rf== nil", rf.me)
	} else {
		if rf.Log.Entries == nil {
			// log.Printf("S[%d] Eh? rf.Log.Entries == nil", rf.me)
		} else {
			// log.Printf("S[%d] #len(rf.log.Entries)=%v rf.CurrentTerm=%v rf.VotedFor=%v",
			// rf.me, len(rf.Log.Entries), rf.CurrentTerm, rf.VotedFor)
		}
	}

	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)

	if e.Encode(rf.CurrentTerm) != nil ||
		e.Encode(rf.VotedFor) != nil ||
		e.Encode(rf.Log) != nil {
		log.Fatalf("failed to stateEncode")
	}

	return w.Bytes()
}

func (rf *Raft) persist() {
	data := rf.stateEncode()
	rf.persister.SaveRaftState(data)
}

func (rf *Raft) RaftStateSize() int {
	return rf.persister.RaftStateSize()
}

func (rf *Raft) persistStateAndSnapShot(snapshot []byte) {
	data := rf.stateEncode()
	rf.persister.SaveStateAndSnapshot(data, snapshot)
}

// 恢复之前持久化的状态。
func (rf *Raft) readPersist(data []byte) {
	defer func() {
		log.Printf("S[%03d] [INIT] len(log)=%+v CurrentTerm=%v VotedFor=%v\n",
			rf.me, rf.Log.Len(), rf.CurrentTerm, rf.VotedFor)
	}()

	if data == nil || len(data) < 1 { // bootstrap without any state?
		rf.Log.Entries = append(rf.Log.Entries, RaftLog{Command: "INVALID"}) /* 空项 */
		rf.Log.LastIncludedIndex = -1
		rf.Log.LastIncludedTerm = -1
		rf.CurrentTerm = 0
		rf.VotedFor = -1
		rf.Init()
		return
	}

	var Logs RaftLogs
	var CurrentTerm int
	var VotedFor int

	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	if (d.Decode(&CurrentTerm) != nil) ||
		(d.Decode(&VotedFor) != nil) ||
		(d.Decode(&Logs) != nil) {
		log.Fatalf("failed to readPersist")
	} else {
		rf.CurrentTerm = CurrentTerm
		rf.VotedFor = VotedFor
		rf.commitIndex = -1
		rf.lastApplied = -1
		if Logs.Entries == nil {
			rf.Log.Entries = make([]RaftLog, 0)
		} else {
			rf.Log.Entries = Logs.Entries
		}
		if Logs.LastIncludedIndex != -1 {
			rf.ReStoreSnapshot(Logs.LastIncludedIndex, Logs.LastIncludedTerm)
		} else {
			rf.Init()
		}
	}
}

func (rf *Raft) ReStoreSnapshot(lastIncludedIndex int, lastIncludedTerm int) {
	rf.ReadSnapshot(rf.persister.ReadSnapshot(), lastIncludedIndex, lastIncludedTerm)
}

func (rf *Raft) ReadSnapshot(data []byte, lastIncludedIndex int, lastIncludedTerm int) {
	rf.mu.Lock()
	msg := &ApplyMsg{
		SnapshotValid: true,
		Snapshot:      data,
		SnapshotIndex: lastIncludedIndex,
		SnapshotTerm:  lastIncludedTerm,
	}
	rf.mu.Unlock()
	go func() { rf.applyCh <- *msg }()
}
