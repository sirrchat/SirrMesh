package pass_blockchain

import (
	"context"
	"strings"

	"github.com/sirrchat/SirrMesh/framework/address"
	parser "github.com/sirrchat/SirrMesh/framework/cfgparser"
	"github.com/sirrchat/SirrMesh/framework/config"
	modconfig "github.com/sirrchat/SirrMesh/framework/config/module"
	"github.com/sirrchat/SirrMesh/framework/log"
	"github.com/sirrchat/SirrMesh/framework/module"
)

type EVMAuth struct {
	modName    string
	instName   string
	inlineArgs []string

	log log.Logger
	// custom fields
	chain   module.BlockChain
	storage module.ManageableStorage
}

func NewEVM(modName, instName string, _, inlineArgs []string) (module.Module, error) {
	return &EVMAuth{
		modName:    modName,
		instName:   instName,
		inlineArgs: inlineArgs,
		log:        log.Logger{Name: "auth.pass_evm"},
	}, nil
}

func (a *EVMAuth) Init(cfg *config.Map) error {
	err := modconfig.ModuleFromNode("blockchain", cfg.Block.Children[0].Args, parser.Node{}, cfg.Globals, &a.chain)
	if err != nil {
		a.log.Printf("error initializing blockchain: %v", err)
		return err
	}
	err = modconfig.ModuleFromNode("storage", cfg.Block.Children[1].Args, parser.Node{}, cfg.Globals, &a.storage)
	return nil
}

func (a *EVMAuth) Name() string {
	return a.modName
}

func (a *EVMAuth) InstanceName() string {
	return a.instName
}

func (a *EVMAuth) AuthPlain(username, sign string) error {
	pk, _, err := address.Split(username)
	if err != nil {
		a.log.Printf("error splitting address: %v", err)
		return err
	}
	result, err := a.chain.CheckSign(context.TODO(), pk, sign, strings.ToLower(pk))
	if err != nil {
		a.log.Printf("error checking signature: %v", err)
		return err
	}
	if !result { // signature is not valid
		return module.ErrUnknownCredentials
	}
	// check if the user not exists in the storage and create it
	//user, err := a.storage.GetIMAPAcct(username)
	//if err == nil && user == nil {
	//	_ = a.storage.CreateIMAPAcct(username)
	//	user, _ = a.storage.GetIMAPAcct(username)
	//	if user != nil {
	//		var errs []error
	//		errs = append(errs, user.CreateMailbox(imap.SentAttr))
	//		errs = append(errs, user.CreateMailbox(imap.TrashAttr))
	//		errs = append(errs, user.CreateMailbox(imap.JunkAttr))
	//		errs = append(errs, user.CreateMailbox(imap.DraftsAttr))
	//		errs = append(errs, user.CreateMailbox(imap.ArchiveAttr))
	//		for _, e := range errs {
	//			if e != nil {
	//				a.log.Printf("error creating mailbox: %v", e)
	//			}
	//		}
	//	}
	//}
	return nil
}

func (a *EVMAuth) ListUsers() ([]string, error) {
	//TODO implement me
	//panic("implement me")
	return a.storage.ListIMAPAccts()
}

func (a *EVMAuth) CreateUser(username, password string) error {
	return a.storage.CreateIMAPAcct(username)
}

func (a *EVMAuth) SetUserPassword(username, password string) error {
	return nil
}

func (a *EVMAuth) DeleteUser(username string) error {
	return a.storage.DeleteIMAPAcct(username)
}

func init() {
	module.Register("auth.pass_evm", NewEVM)
}
