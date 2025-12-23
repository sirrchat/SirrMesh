/*
SirrMesh - Composable all-in-one email server.
Copyright © 2019-2020 Max Mazurov <fox.cpp@disroot.org>, SirrMesh contributors

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/term"

	"github.com/mail-chat-chain/sirrmeshd/framework/config"
	"github.com/mail-chat-chain/sirrmeshd/framework/log"
)

/*
Config matchers for module interfaces.
*/

var stdinScanner = bufio.NewScanner(os.Stdin)

// logOut structure wraps log.Output and preserves
// configuration directive it was constructed from, allowing
// dynamic reinitialization for purposes of log file rotation.
type logOut struct {
	args []string
	log.Output
}

func logOutput(_ *config.Map, node config.Node) (interface{}, error) {
	if len(node.Args) == 0 {
		return nil, config.NodeErr(node, "expected at least 1 argument")
	}
	if len(node.Children) != 0 {
		return nil, config.NodeErr(node, "can't declare block here")
	}

	return LogOutputOption(node.Args)
}

func LogOutputOption(args []string) (log.Output, error) {
	outs := make([]log.Output, 0, len(args))
	for i, arg := range args {
		switch arg {
		case "stderr":
			outs = append(outs, log.WriterOutput(os.Stderr, false))
		case "stderr_ts":
			outs = append(outs, log.WriterOutput(os.Stderr, true))
		case "syslog":
			syslogOut, err := log.SyslogOutput()
			if err != nil {
				return nil, fmt.Errorf("failed to connect to syslog daemon: %v", err)
			}
			outs = append(outs, syslogOut)
		case "off":
			if len(args) != 1 {
				return nil, errors.New("'off' can't be combined with other log targets")
			}
			return log.NopOutput{}, nil
		default:
			// Log file paths are converted to absolute to make sure
			// we will be able to recreate them in right location
			// after changing working directory to the state dir.
			absPath, err := filepath.Abs(arg)
			if err != nil {
				return nil, err
			}
			// We change the actual argument, so logOut object will
			// keep the absolute path for reinitialization.
			args[i] = absPath

			w, err := os.OpenFile(absPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o666)
			if err != nil {
				return nil, fmt.Errorf("failed to create log file: %v", err)
			}

			outs = append(outs, log.WriteCloserOutput(w, true))
		}
	}

	if len(outs) == 1 {
		return logOut{args, outs[0]}, nil
	}
	return logOut{args, log.MultiOutput(outs...)}, nil
}

func defaultLogOutput() (interface{}, error) {
	return log.DefaultLogger.Out, nil
}

func reinitLogging() {
	out, ok := log.DefaultLogger.Out.(logOut)
	if !ok {
		log.Println("Can't reinitialize logger because it was replaced before, this is a bug")
		return
	}

	newOut, err := LogOutputOption(out.args)
	if err != nil {
		log.Println("Can't reinitialize logger:", err)
		return
	}

	out.Close()

	log.DefaultLogger.Out = newOut
}

// generateMailConfigContent generates the default sirrmeshd.conf content from project template
func generateMailConfigContent() string {
	return `## SirrMesh - default configuration file (2022-06-18)
# Suitable for small-scale deployments. Uses its own format for local users DB,
# should be managed via sirrmeshd subcommands.
#
# See tutorials at https://maddy.email for guidance on typical
# configuration changes.

# ----------------------------------------------------------------------------
# Base variables

$(hostname) = example.com
$(primary_domain) = example.com
$(local_domains) = $(primary_domain)

# tls file certs/$(hostname)/fullchain.pem certs/$(hostname)/privkey.pem

tls {
    loader acme {
        hostname $(hostname)
        email postmaster@$(hostname)
        agreed
        challenge dns-01
        dns cloudflare {
            api_token CLOUDFLARE_API_TOKEN
        }
    }
}

# ----------------------------------------------------------------------------
# blockchains
# blockchain.ethereum sirrmeshd {
#     chain_id 26000
#     rpc_url http://127.0.0.1:8545
# }
blockchain.ethereum bsc {
    chain_id 56
    rpc_url https://binance.llamarpc.com
}

# ----------------------------------------------------------------------------
# Local storage & authentication

# imapsql module stores all indexes and metadata necessary for IMAP using a
# relational database. It is used by IMAP endpoint for mailbox access and
# also by SMTP & Submission endpoints for delivery of local messages.
#
# IMAP accounts, mailboxes and all message metadata can be inspected using
# imap-* subcommands of sirrmeshd.

storage.imapsql local_mailboxes {
    driver sqlite3
    dsn imapsql.db
}

# pass_table provides local hashed passwords storage for authentication of
# users. It can be configured to use any "table" module, in default
# configuration a table in SQLite DB is used.
# Table can be replaced to use e.g. a file for passwords. Or pass_table module
# can be replaced altogether to use some external source of credentials (e.g.
# PAM, /etc/shadow file).
#
# If table module supports it (sql_table does) - credentials can be managed
# using 'sirrmeshd creds' command.

# auth.pass_table local_authdb {
#     table sql_table {
#         driver sqlite3
#         dsn credentials.db
#         table_name passwords
#     }
# }

# pass blockchain module provides authentication using blockchain wallets.
auth.pass_evm blockchain_atuh {
    blockchain &bsc
    storage &local_mailboxes
}

# ----------------------------------------------------------------------------
# SMTP endpoints + message routing

hostname $(hostname)

table.chain local_rewrites {
    optional_step regexp "(.+)\+(.+)@(.+)" "$1@$3"
    optional_step static {
        entry postmaster postmaster@$(primary_domain)
    }
    optional_step file ~/.mailcoin/aliases
}

msgpipeline local_routing {
    # Insert handling for special-purpose local domains here.
    # e.g.
    # destination lists.example.org {
    #     deliver_to lmtp tcp://127.0.0.1:8024
    # }

    destination postmaster $(local_domains) {
        modify {
            replace_rcpt &local_rewrites
            blockchain_tx &bsc
        }

        deliver_to &local_mailboxes
    }

    default_destination {
        reject 550 5.1.1 "User doesn't exist"
    }
}

smtp tcp://0.0.0.0:8825 {
    limits {
        # Up to 20 msgs/sec across max. 10 SMTP connections.
        all rate 20 1s
        all concurrency 10
    }

    dmarc yes
    check {
        require_mx_record
        dkim
        spf
    }

    source $(local_domains) {
        reject 501 5.1.8 "Use Submission for outgoing SMTP"
    }
    default_source {
        destination postmaster $(local_domains) {
            deliver_to &local_routing
        }
        default_destination {
            reject 550 5.1.1 "User doesn't exist"
        }
    }
}

submission tls://0.0.0.0:465 tcp://0.0.0.0:587 {
    limits {
        # Up to 50 msgs/sec across any amount of SMTP connections.
        all rate 50 1s
    }

    auth &blockchain_atuh

    source $(local_domains) {
        check {
            authorize_sender {
                prepare_email &local_rewrites
                user_to_email identity
            }
        }

        modify {
            blockchain_tx &bsc
        }

        destination postmaster $(local_domains) {
            deliver_to &local_routing
        }
        default_destination {
            modify {
                dkim $(primary_domain) $(local_domains) default
            }
            deliver_to &remote_queue
        }
    }
    default_source {
        reject 501 5.1.8 "Non-local sender domain"
    }
}

target.remote outbound_delivery {
    # Set outbound SMTP connection port (default is 25)
    smtp_port 8825    # Set to 8825
    limits {
        # Up to 20 msgs/sec across max. 10 SMTP connections
        # for each recipient domain.
        destination rate 20 1s
        destination concurrency 10
    }
    mx_auth {
        dane {
            # SMTP port for DANE TLSA record queries (should match smtp_port above)
            # Uncomment and modify if using custom port
            smtp_port 8825
        }
        mtasts {
            cache fs
            fs_dir mtasts_cache/
        }
        local_policy {
            min_tls_level encrypted
            min_mx_level none
        }
    }
}

target.queue remote_queue {
    target &outbound_delivery

    autogenerated_msg_domain $(primary_domain)
    bounce {
        destination postmaster $(local_domains) {
            deliver_to &local_routing
        }
        default_destination {
            reject 550 5.0.0 "Refusing to send DSNs to non-local addresses"
        }
    }
}

# ----------------------------------------------------------------------------
# IMAP endpoints

imap tls://0.0.0.0:993 tcp://0.0.0.0:143 {
    auth &blockchain_atuh
	storage &local_mailboxes
}
`
}

// ReadPassword prompts the user for a password with hidden input
func ReadPassword(prompt string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", prompt)

	// Check if stdin is a terminal
	if !term.IsTerminal(int(syscall.Stdin)) {
		// If not a terminal, read plaintext (for scripts/tests)
		if !stdinScanner.Scan() {
			return "", stdinScanner.Err()
		}
		return stdinScanner.Text(), nil
	}

	// Read password with hidden input
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr) // Print newline after password input

	if err != nil {
		return "", err
	}

	return string(password), nil
}
