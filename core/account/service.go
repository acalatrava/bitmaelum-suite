package account

import (
    "errors"
    "github.com/jaytaph/mailv2/core"
    "github.com/jaytaph/mailv2/core/messagebox"
)

type Service struct {
    repo Repository
}

// Create new service
func AccountService(repo Repository) *Service {
    return &Service{
        repo: repo,
    }
}

// Create new account for the given address and public key
func (s *Service) CreateAccount(addr core.HashAddress, pubKey string) error {
    if s.repo.Exists(addr) {
        return errors.New("account already exists")
    }

    err := s.repo.Create(addr)
    if err != nil {
        return err
    }

    _ = s.repo.CreateBox(addr, "inbox", "This is your regular inbox", 0)
    _ = s.repo.CreateBox(addr, "outbox", "All your outgoing messages will be stored here", 0)
    _ = s.repo.CreateBox(addr, "trash", "Trashcan. Everything in here will be removed automatically after 30 days or when purged manually", 0)
    _ = s.repo.StorePubKey(addr, pubKey)

    return nil
}

// Check if account exists for address
func (s *Service) AccountExists(addr core.HashAddress) bool {
    return s.repo.Exists(addr)
}

// Retrieve the public keys for given address
func (s *Service) GetPublicKeys(addr core.HashAddress) []string {
    if ! s.repo.Exists(addr) {
        return []string{}
    }

    pubKeys, err := s.repo.FetchPubKeys(addr)
    if err != nil {
        return []string{}
    }

    return pubKeys
}

func (s *Service) FetchMessageBoxes(addr core.HashAddress, query string) []messagebox.MailBoxInfo {
    list, err := s.repo.FindBox(addr, query)
    if err != nil {
        return []messagebox.MailBoxInfo{}
    }

    return list
}

func (s *Service) FetchMailFromBox(addr core.HashAddress, box string, offset int, limit int) []messagebox.MessageInfo {
    list, err := s.repo.FindMessages(addr, box, offset, limit)
    if err != nil {
        return []messagebox.MessageInfo{}
    }

    return list
}
