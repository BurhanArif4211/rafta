package sync

import (
    "encoding/json"
    "log"
    "github.com/burhanarif4211/rafta/internal/models"
		"github.com/burhanarif4211/rafta/internal/repository"
    "net/http"
    "time"
)

type Server struct {
    noteFolderRepo repository.NoteFolderRepository
    noteRepo       repository.NoteRepository
    todoFolderRepo repository.TodoFolderRepository
    todoRepo       repository.TodoRepository
    todoStepRepo   repository.TodoStepRepository
    httpServer     *http.Server
}

func NewServer(
    nfRepo repository.NoteFolderRepository,
    nRepo repository.NoteRepository,
    tfRepo repository.TodoFolderRepository,
    tRepo repository.TodoRepository,
    tsRepo repository.TodoStepRepository,
) *Server {
    return &Server{
        noteFolderRepo: nfRepo,
        noteRepo:       nRepo,
        todoFolderRepo: tfRepo,
        todoRepo:       tRepo,
        todoStepRepo:   tsRepo,
    }
}

func (s *Server) Start(port string) error {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/sync", s.handleSync)

    s.httpServer = &http.Server{
        Addr:         ":" + port,
        Handler:      mux,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    log.Printf("Sync server starting on port %s", port)
    // Run in goroutine; caller should handle error (e.g., log.Fatal if critical)
    go func() {
        if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Printf("Sync server error: %v", err)
        }
    }()
    return nil
}

func (s *Server) Stop() error {
    if s.httpServer != nil {
        return s.httpServer.Close()
    }
    return nil
}

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Fetch all data from repositories
    noteFolders, err := s.noteFolderRepo.GetAll()
    if err != nil {
        http.Error(w, "failed to fetch note folders", http.StatusInternalServerError)
        return
    }
    notes, err := s.noteRepo.GetAll()
    if err != nil {
        http.Error(w, "failed to fetch notes", http.StatusInternalServerError)
        return
    }
    todoFolders, err := s.todoFolderRepo.GetAll()
    if err != nil {
        http.Error(w, "failed to fetch todo folders", http.StatusInternalServerError)
        return
    }
    todos, err := s.todoRepo.GetAll()
    if err != nil {
        http.Error(w, "failed to fetch todos", http.StatusInternalServerError)
        return
    }
    todoSteps, err := s.todoStepRepo.GetAll()
    if err != nil {
        http.Error(w, "failed to fetch todo steps", http.StatusInternalServerError)
        return
    }

    // Convert to sync DTOs
    syncData := SyncData{
        NoteFolders: toNoteFolderSyncList(noteFolders),
        Notes:       toNoteSyncList(notes),
        TodoFolders: toTodoFolderSyncList(todoFolders),
        Todos:       toTodoSyncList(todos),
        TodoSteps:   toTodoStepSyncList(todoSteps),
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(syncData); err != nil {
        log.Printf("failed to encode sync data: %v", err)
    }
}

// Conversion helpers (internal to sync)
func toNoteFolderSyncList(folders []*models.NoteFolder) []NoteFolderSync {
    var out []NoteFolderSync
    for _, f := range folders {
        var parent *string
        if f.ParentID.Valid {
            parent = &f.ParentID.String
        }
        out = append(out, NoteFolderSync{
            ID:        f.ID,
            Name:      f.Name,
            ParentID:  parent,
            CreatedAt: f.CreatedAt,
            UpdatedAt: f.UpdatedAt,
        })
    }
    return out
}

func toNoteSyncList(notes []*models.Note) []NoteSync {
    var out []NoteSync
    for _, n := range notes {
        out = append(out, NoteSync{
            ID:        n.ID,
            Title:     n.Title,
            Content:   n.Content,
            FolderID:  n.FolderID,
            CreatedAt: n.CreatedAt,
            UpdatedAt: n.UpdatedAt,
        })
    }
    return out
}

func toTodoFolderSyncList(folders []*models.TodoFolder) []TodoFolderSync {
    var out []TodoFolderSync
    for _, f := range folders {
        var parent *string
        if f.ParentID.Valid {
            parent = &f.ParentID.String
        }
        out = append(out, TodoFolderSync{
            ID:        f.ID,
            Name:      f.Name,
            ParentID:  parent,
            CreatedAt: f.CreatedAt,
            UpdatedAt: f.UpdatedAt,
        })
    }
    return out
}

func toTodoSyncList(todos []*models.Todo) []TodoSync {
    var out []TodoSync
    for _, t := range todos {
        out = append(out, TodoSync{
            ID:        t.ID,
            Title:     t.Title,
            FolderID:  t.FolderID,
            CreatedAt: t.CreatedAt,
            UpdatedAt: t.UpdatedAt,
        })
    }
    return out
}

func toTodoStepSyncList(steps []*models.TodoStep) []TodoStepSync {
    var out []TodoStepSync
    for _, s := range steps {
        out = append(out, TodoStepSync{
            ID:          s.ID,
            TodoID:      s.TodoID,
            Description: s.Description,
            Completed:   s.Completed,
            DisplayOrder: s.DisplayOrder,
            CreatedAt:   s.CreatedAt,
            UpdatedAt:   s.UpdatedAt,
        })
    }
    return out
}
