package repository

import (
	"mqtt-streaming-server/domain"
	"mqtt-streaming-server/utils"
)

func EncryptPhotoFields(photo *domain.Photo) error {
	var err error
	photo.Text, err = utils.EncryptString(photo.Text)
	if err != nil {
		return err
	}
	photo.UnitateMedicala, err = utils.EncryptString(photo.UnitateMedicala)
	if err != nil {
		return err
	}
	photo.AdresaUnitateMedicala, err = utils.EncryptString(photo.AdresaUnitateMedicala)
	if err != nil {
		return err
	}
	photo.TelefonUnitateMedicala, err = utils.EncryptString(photo.TelefonUnitateMedicala)
	if err != nil {
		return err
	}
	photo.NumarFisa, err = utils.EncryptString(photo.NumarFisa)
	if err != nil {
		return err
	}
	photo.SocietateUnitate, err = utils.EncryptString(photo.SocietateUnitate)
	if err != nil {
		return err
	}
	photo.AdresaAngajator, err = utils.EncryptString(photo.AdresaAngajator)
	if err != nil {
		return err
	}
	photo.TelefonAngajator, err = utils.EncryptString(photo.TelefonAngajator)
	if err != nil {
		return err
	}
	photo.Nume, err = utils.EncryptString(photo.Nume)
	if err != nil {
		return err
	}
	photo.Prenume, err = utils.EncryptString(photo.Prenume)
	if err != nil {
		return err
	}
	photo.CNP, err = utils.EncryptString(photo.CNP)
	if err != nil {
		return err
	}
	photo.ProfesieFunctie, err = utils.EncryptString(photo.ProfesieFunctie)
	if err != nil {
		return err
	}
	photo.LocDeMunca, err = utils.EncryptString(photo.LocDeMunca)
	if err != nil {
		return err
	}
	photo.Recomandari, err = utils.EncryptString(photo.Recomandari)
	if err != nil {
		return err
	}
	return nil
}

func DecryptPhotoFields(photo *domain.Photo) error {
	var err error
	photo.Text, err = utils.DecryptString(photo.Text)
	if err != nil {
		return err
	}
	photo.UnitateMedicala, err = utils.DecryptString(photo.UnitateMedicala)
	if err != nil {
		return err
	}
	photo.AdresaUnitateMedicala, err = utils.DecryptString(photo.AdresaUnitateMedicala)
	if err != nil {
		return err
	}
	photo.TelefonUnitateMedicala, err = utils.DecryptString(photo.TelefonUnitateMedicala)
	if err != nil {
		return err
	}
	photo.NumarFisa, err = utils.DecryptString(photo.NumarFisa)
	if err != nil {
		return err
	}
	photo.SocietateUnitate, err = utils.DecryptString(photo.SocietateUnitate)
	if err != nil {
		return err
	}
	photo.AdresaAngajator, err = utils.DecryptString(photo.AdresaAngajator)
	if err != nil {
		return err
	}
	photo.TelefonAngajator, err = utils.DecryptString(photo.TelefonAngajator)
	if err != nil {
		return err
	}
	photo.Nume, err = utils.DecryptString(photo.Nume)
	if err != nil {
		return err
	}
	photo.Prenume, err = utils.DecryptString(photo.Prenume)
	if err != nil {
		return err
	}
	photo.CNP, err = utils.DecryptString(photo.CNP)
	if err != nil {
		return err
	}
	photo.ProfesieFunctie, err = utils.DecryptString(photo.ProfesieFunctie)
	if err != nil {
		return err
	}
	photo.LocDeMunca, err = utils.DecryptString(photo.LocDeMunca)
	if err != nil {
		return err
	}
	photo.Recomandari, err = utils.DecryptString(photo.Recomandari)
	if err != nil {
		return err
	}
	return nil
}
